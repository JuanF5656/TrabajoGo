"use strict";

// Etiquetas en español y color de cada valor que manda el backend.
const STATE_LABEL = {
  Awake: ["Despierto", "b-ok"],
  AsleepInHallway: ["Dormido en pasillo", "b-bad"],
  AsleepInBed: ["Dormido en cama", "b-info"],
};
const LEVEL_CLASS = { Mild: "b-ok", Moderate: "b-warn", Severe: "b-bad" };
const ROOM_LABEL = { Available: ["Disponible", "b-ok"], Occupied: ["Ocupada", "b-bad"] };

// Lo que el usuario escribió en cada campo de ubicación sobrevive a los re-dibujos.
const typedLocations = new Map();
let noticeTimer = null;

// Crea un elemento. Todo el texto entra con textContent (nunca innerHTML),
// así un nombre con "<" no puede inyectar HTML.
function el(tag, opts = {}, children = []) {
  const node = document.createElement(tag);
  if (opts.cls) node.className = opts.cls;
  if (opts.text !== undefined) node.textContent = opts.text;
  for (const [k, v] of Object.entries(opts.attrs || {})) node.setAttribute(k, v);
  for (const c of children) node.append(c);
  return node;
}

function badge(label, cls) {
  return el("span", { cls: "badge " + cls, text: label });
}

// Llama al backend. Si responde con error HTTP, lanza Error con el mensaje del JSON.
async function api(method, path, body) {
  const res = await fetch(path, {
    method,
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
  let data = null;
  try { data = await res.json(); } catch (_) { /* respuesta sin cuerpo */ }
  if (!res.ok) throw new Error((data && data.error) || "Error " + res.status);
  return data;
}

function showNotice(message, kind) {
  const box = document.getElementById("notice");
  box.textContent = message;
  box.className = "notice " + kind;
  box.hidden = false;
  clearTimeout(noticeTimer);
  noticeTimer = setTimeout(() => { box.hidden = true; }, 6000);
}

// Ejecuta una acción, muestra el aviso y SIEMPRE recarga estado y consultas
// (una asignación fallida también cambia la consulta 5.3).
async function act(method, path, body, okMessage) {
  try {
    const data = await api(method, path, body);
    if (data && data.warning) showNotice(data.warning, "warn");
    else showNotice(okMessage(data), "ok");
  } catch (err) {
    showNotice(err.message, "error");
  }
  await refresh();
}

// ---------- dibujo ----------

function renderRooms(rooms) {
  const box = document.getElementById("rooms");
  box.replaceChildren(...rooms.map((r) => {
    const [label, cls] = ROOM_LABEL[r.state] || [r.state, "b-gray"];
    const occupant = r.occupants.length ? r.occupants.join(", ") : "Vacía";
    return el("div", { cls: "card" }, [
      el("div", { cls: "num", text: "Sala " + r.number }),
      el("div", { cls: "occ", text: occupant }),
      badge(label + " (" + r.occupants.length + "/" + r.capacity + ")", cls),
    ]);
  }));
}

function renderPatients(patients) {
  const body = document.getElementById("patients");
  body.replaceChildren(...patients.map((p) => {
    const [stLabel, stCls] = STATE_LABEL[p.state] || [p.state, "b-gray"];

    const input = el("input", { attrs: { type: "text", placeholder: "¿Dónde le dio el ataque?", "aria-label": "Ubicación de " + p.name } });
    input.value = typedLocations.get(p.id) || "";
    input.addEventListener("input", () => typedLocations.set(p.id, input.value));

    const attack = el("button", { text: "Ataque de sueño", attrs: { type: "button" } });
    attack.disabled = p.state !== "Awake";
    attack.addEventListener("click", () => {
      const location = input.value.trim();
      typedLocations.delete(p.id);
      act("POST", "/api/attack", { patientId: p.id, location },
        (d) => d.episode);
    });

    const wake = el("button", { text: "Despertar", cls: "secondary", attrs: { type: "button" } });
    wake.disabled = p.state === "Awake";
    wake.addEventListener("click", () =>
      act("POST", "/api/wakeup", { patientId: p.id }, (d) => d.message));

    const assign = el("button", { text: "Asignar sala", cls: "secondary", attrs: { type: "button" } });
    assign.disabled = p.state !== "AsleepInHallway";
    assign.addEventListener("click", () =>
      act("POST", "/api/assign", { patientId: p.id }, (d) => d.message));

    return el("tr", {}, [
      el("td", { text: p.id }),
      el("td", { text: p.name }),
      el("td", { text: String(p.age) }),
      el("td", {}, [badge(p.level, LEVEL_CLASS[p.level] || "b-gray")]),
      el("td", {}, [badge(stLabel, stCls)]),
      el("td", { text: p.location || "—" }),
      el("td", { text: p.room ? String(p.room) : "—" }),
      el("td", {}, [
        el("div", { cls: "actions" }, [
          el("div", { cls: "attack" }, [input, attack]),
          wake,
          assign,
        ]),
      ]),
    ]);
  }));
}

function renderStaff(staff) {
  const list = document.getElementById("staff");
  list.replaceChildren(...staff.map((s) =>
    el("li", {}, [
      el("span", { text: s.name + " (" + s.id + ")" }),
      badge(s.type, s.type === "Doctor" ? "b-info" : "b-gray"),
    ])));
}

function section(title, content) {
  return el("div", { cls: "query" }, [el("h3", { text: title }), ...content]);
}

function listOrEmpty(items, emptyText) {
  if (!items.length) return [el("p", { cls: "empty", text: emptyText })];
  return [el("ul", {}, items.map((t) => el("li", { text: t })))];
}

function renderQueries(q) {
  const box = document.getElementById("queries");
  const parts = [];

  parts.push(section("5.1 Pacientes dormidos en el pasillo",
    listOrEmpty(q.hallway.map((p) => p.id + " " + p.name + " (" + p.level + ") en " + p.location),
      "Nadie está dormido en el pasillo.")));

  for (const d of q.episodesByDoctor) {
    parts.push(section("5.2 Episodios atendidos por " + d.doctorName,
      listOrEmpty(d.episodes, "Sin episodios todavía.")));
  }

  const last = q.lastAssignment;
  parts.push(section("5.3 Asignación de sala" + (last ? " a " + last.patientId : ""),
    last
      ? [el("p", {}, [badge(last.ok ? "Éxito" : "Sin sala", last.ok ? "b-ok" : "b-bad"), " " + last.message])]
      : [el("p", { cls: "empty", text: "Todavía no se ha asignado ninguna sala." })]));

  parts.push(section("5.4 Reporte de severidad",
    listOrEmpty(q.severeReport.map((s) => s.id + " " + s.name + ": " + s.episodes + " episodio(s)"),
      "No hay pacientes Severe.")));

  box.replaceChildren(...parts);
}

async function refresh() {
  try {
    const [state, queries] = await Promise.all([api("GET", "/api/state"), api("GET", "/api/queries")]);
    document.getElementById("hospital-name").textContent = state.hospital;
    document.title = state.hospital;
    renderRooms(state.rooms);
    renderPatients(state.patients);
    renderStaff(state.staff);
    renderQueries(queries);
  } catch (err) {
    showNotice("No se pudo cargar el estado: " + err.message, "error");
  }
}

document.getElementById("reset-btn").addEventListener("click", () => {
  typedLocations.clear();
  act("POST", "/api/reset", null, (d) => d.message);
});

refresh();
