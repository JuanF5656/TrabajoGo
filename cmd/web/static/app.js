"use strict";

/* =====================================================================
   HOSPITAL DE LOS COSTEÑOS CON NARCOLEPSIA — interfaz web

   Esta página es SOLO la vista: no tiene reglas de negocio. Toda la lógica
   (quién se duerme, qué sala se asigna, cuántos episodios hay) la decide el
   paquete Go "hospital"; aquí solo se pide por HTTP, se dibuja y se anima.

   Flujo general:  acción del usuario -> Actions (POST) -> refresh() (GET)
                   -> Store (estado) -> módulos de dibujo (Plan, Panel, ...)

   Módulos de este archivo, en orden:
     1. Util ........ funciones pequeñas (crear elementos, íconos, textos)
     2. Store/Api ... estado en memoria del navegador y llamadas fetch
     3. Toast ....... avisos animados (éxito / error / aviso)
     4. Plan ........ el plano SVG/CSS: zonas, camas y avatares animados
     5. Panel ....... detalle del paciente seleccionado y lista de personal
     6. Actions ..... ataque, despertar y asignar sala (hablan con el backend)
     7. Interact .... clics, doble clic y arrastrar-y-soltar (drag & drop)
     8. FX .......... contadores animados y confeti
     9. Carnival .... modo simulación: ataques aleatorios desde el navegador
    10. Arranque .... refresh() y conexión de los botones
   ===================================================================== */

/* =====================================================================
   1. UTIL — funciones pequeñas que usan todos los módulos
   ===================================================================== */
const SVG_NS = "http://www.w3.org/2000/svg";

const $ = (selector, root = document) => root.querySelector(selector);

// Crea un elemento HTML. Todo texto entra con textContent (nunca innerHTML):
// así un nombre como "<img onerror=...>" se muestra como texto y no se ejecuta.
function el(tag, opts = {}, children = []) {
  const node = document.createElement(tag);
  if (opts.cls) node.className = opts.cls;
  if (opts.text !== undefined) node.textContent = opts.text;
  for (const [k, v] of Object.entries(opts.attrs || {})) node.setAttribute(k, v);
  for (const c of children) node.append(c);
  return node;
}

// Ícono SVG: apunta a un <symbol id="i-nombre"> del "sprite" de index.html.
function icon(name, cls = "") {
  const svg = document.createElementNS(SVG_NS, "svg");
  svg.setAttribute("class", ("ico " + cls).trim());
  svg.setAttribute("aria-hidden", "true");
  svg.setAttribute("focusable", "false");
  const use = document.createElementNS(SVG_NS, "use");
  use.setAttribute("href", "#i-" + name);
  svg.append(use);
  return svg;
}

const reducedMotion = () => window.matchMedia("(prefers-reduced-motion: reduce)").matches;
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const rand = (min, max) => min + Math.random() * (max - min);
const pick = (list) => list[Math.floor(Math.random() * list.length)];

// "Cafetería" -> "cafeteria": sirve para comparar nombres sin tildes ni mayúsculas.
const normalize = (s) => String(s || "").normalize("NFD").replace(/[̀-ͯ]/g, "").toLowerCase().trim();

// "Yeimy Padilla" -> "YP"
function initials(name) {
  const words = String(name).trim().split(/\s+/).filter(Boolean);
  if (!words.length) return "?";
  const first = words[0][0];
  const last = words.length > 1 ? words[words.length - 1][0] : "";
  return (first + last).toUpperCase();
}
const firstName = (name) => String(name).trim().split(/\s+/)[0] || "";

// localStorage puede fallar (ventana privada, bloqueado): siempre dentro de try/catch.
const Safe = {
  get(key) { try { return localStorage.getItem(key); } catch (_) { return null; } },
  set(key, value) { try { localStorage.setItem(key, value); } catch (_) { /* no pasa nada */ } },
};

// Textos en español de los valores que manda el backend.
const STATE_LABEL = {
  Awake: "Despierto",
  AsleepInHallway: "Dormido en pasillo",
  AsleepInBed: "Dormido en cama",
};
const STATE_BADGE = { Awake: "b-ok", AsleepInHallway: "b-bad", AsleepInBed: "b-info" };

// Las zonas del plano. "name" es lo que se manda al backend como ubicación.
const ZONES = [
  { key: "cafeteria", name: "Cafetería", icon: "coffee" },
  { key: "recepcion", name: "Recepción", icon: "desk" },
  { key: "radiologia", name: "Radiología", icon: "xray" },
  { key: "biblioteca", name: "Biblioteca", icon: "book" },
  { key: "pasillo", name: "Pasillo principal", icon: "hall" },
];
const zoneByKey = (key) => ZONES.find((z) => z.key === key);
// Dado un texto de ubicación ("Cafetería", "cafeteria"...) devuelve la zona, o null.
function zoneFromLocation(text) {
  const n = normalize(text);
  return ZONES.find((z) => z.key === n || normalize(z.name) === n) || null;
}

/* =====================================================================
   2. STORE y API — el estado vive aquí; el servidor es la fuente de verdad
   ===================================================================== */
const Store = {
  data: null,              // última respuesta de GET /api/state
  queries: null,           // última respuesta de GET /api/queries
  selectedId: null,        // paciente seleccionado (o null)
  zoneOf: new Map(),       // id -> zona donde "camina" el paciente mientras está despierto

  patient(id) { return this.data ? this.data.patients.find((p) => p.id === id) : undefined; },
  awake() { return this.data ? this.data.patients.filter((p) => p.state === "Awake") : []; },

  // El backend no sabe en qué zona está un paciente despierto (no existe ese dato
  // en el dominio), así que la vista le reparte una zona inicial a cada uno.
  seedZones(patients) {
    const walkable = ZONES.slice(0, 4);
    patients.forEach((p, i) => {
      if (!this.zoneOf.has(p.id)) this.zoneOf.set(p.id, walkable[i % walkable.length].key);
    });
  },

  // Se llama al reiniciar el escenario: el cliente olvida lo que había aprendido.
  reset() {
    this.selectedId = null;
    this.zoneOf.clear();
  },
};

// Llama al backend. Si responde con error HTTP lanza Error con el mensaje del JSON
// ({"error": "..."}), así quien llama solo tiene un try/catch.
const Api = {
  async request(method, path, body) {
    const res = await fetch(path, {
      method,
      headers: body ? { "Content-Type": "application/json" } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    });
    let data = null;
    try { data = await res.json(); } catch (_) { /* respuesta sin cuerpo */ }
    if (!res.ok) throw new Error((data && data.error) || "Error " + res.status);
    return data;
  },
  get(path) { return this.request("GET", path); },
  post(path, body) { return this.request("POST", path, body || null); },
};

/* =====================================================================
   3. TOAST — avisos animados en una esquina (reemplazan a alert())
   kind: "ok" (verde) | "error" (coral) | "warn" (amarillo) | "info" (azul)
   ===================================================================== */
const Toast = {
  MAX: 4,

  show(kind, message, ms = 5200) {
    const box = $("#toasts");
    const close = el("button", { cls: "t-close", attrs: { type: "button", "aria-label": "Cerrar aviso" } }, [icon("close")]);
    const iconName = { ok: "check", error: "alert", warn: "alert", info: "info" }[kind] || "info";
    const node = el("div", { cls: "toast t-" + kind }, [icon(iconName), el("span", { cls: "t-msg", text: message }), close]);

    let timer = null;
    const dismiss = () => {
      clearTimeout(timer);
      if (!node.isConnected) return;
      if (reducedMotion()) { node.remove(); return; }
      node.classList.add("leaving");
      node.addEventListener("animationend", () => node.remove(), { once: true });
    };
    close.addEventListener("click", dismiss);
    timer = setTimeout(dismiss, ms);

    box.append(node);
    while (box.children.length > this.MAX) box.firstElementChild.remove();   // no apilar infinitos
  },
};

/* =====================================================================
   4. PLAN — el mapa del hospital
   - Las zonas (Cafetería, Recepción...) y las salas son botones en una
     cuadrícula CSS (grid).
   - Los pacientes son "avatares": botones circulares que viven en una capa
     encima del mapa (#avatars) y se colocan con transform: translate(x, y).
     Como la posición es un transform con transition en CSS, cuando cambia
     la zona de un paciente el navegador anima el recorrido solo.
   - Los avatares se crean UNA vez y después solo se actualizan; si se
     recrearan en cada refresh no habría nada que animar.
   ===================================================================== */
const Plan = {
  root: null,
  layer: null,
  roomsRow: null,
  roomKey: "",
  zoneEls: new Map(),    // clave de zona -> botón
  roomEls: new Map(),    // número de sala -> botón
  avatars: new Map(),    // id de paciente -> botón del avatar

  init() {
    this.root = $("#plan");
    this.layer = $("#avatars");
    for (const z of ZONES) {
      const btn = el("button", {
        cls: "zone z-" + z.key,
        attrs: { type: "button", "data-zone": z.key, "aria-label": "Zona " + z.name },
      }, [el("span", { cls: "zone-label" }, [icon(z.icon), el("span", { text: z.name })])]);
      btn.addEventListener("click", () => Interact.zoneClick(z.key));
      this.zoneEls.set(z.key, btn);
      this.root.insertBefore(btn, this.layer);
    }
    this.roomsRow = el("div", { cls: "rooms-row" });
    this.root.insertBefore(this.roomsRow, this.layer);
    // Si cambia el tamaño del plano (girar el celular, redimensionar), se recolocan los avatares.
    new ResizeObserver(() => this.layout()).observe(this.root);
  },

  // Crea (si hace falta) las salas y actualiza su color / estado.
  syncRooms(rooms) {
    const key = rooms.map((r) => r.number).join(",");
    if (key !== this.roomKey) {
      this.roomKey = key;
      this.roomEls.clear();
      this.roomsRow.replaceChildren();
      for (const r of rooms) {
        const btn = el("button", {
          cls: "zone room free",
          attrs: { type: "button", "data-room": String(r.number) },
        }, [
          el("span", { cls: "zone-label" }, [icon("bed"), el("span", { text: "Sala " + r.number })]),
          el("span", { cls: "bed" }, [icon("bed", "bed-ico")]),
          el("span", { cls: "badge room-chip b-ok", text: "Disponible" }),
        ]);
        btn.addEventListener("click", () => Interact.roomClick(r.number));
        this.roomEls.set(r.number, btn);
        this.roomsRow.append(btn);
      }
    }
    for (const r of rooms) {
      const btn = this.roomEls.get(r.number);
      const free = r.state === "Available";
      btn.classList.toggle("free", free);       // el color cambia con una transición CSS
      btn.classList.toggle("occ", !free);
      const chip = $(".room-chip", btn);
      chip.textContent = free ? "Disponible" : "Ocupada";
      chip.className = "badge room-chip " + (free ? "b-ok" : "b-bad");
      const who = r.occupants.length ? "ocupada por " + r.occupants.join(", ") : "disponible";
      btn.setAttribute("aria-label", `Sala ${r.number}, ${who}`);
    }
  },

  // ¿En qué casilla del plano debe estar este paciente ahora mismo?
  zoneKeyFor(p) {
    if (p.state === "AsleepInBed" && p.room) return "room-" + p.room;
    if (p.state === "AsleepInHallway") {
      // Se queda donde le dio el ataque (la ubicación que guardó el backend).
      const zone = zoneFromLocation(p.location);
      const key = zone ? zone.key : "pasillo";
      Store.zoneOf.set(p.id, key);               // cuando despierte, se levanta ahí mismo
      return key;
    }
    return Store.zoneOf.get(p.id) || "pasillo";
  },

  ensureAvatar(p) {
    let a = this.avatars.get(p.id);
    if (a) return a;
    a = el("button", { cls: "avatar", attrs: { type: "button", "data-id": p.id } }, [
      el("span", { cls: "body" }, [el("span", { cls: "ini", text: initials(p.name) })]),
      el("span", { cls: "zzz", attrs: { "aria-hidden": "true" } }, ["Z", "z", "z"].map((t) => el("i", { text: t }))),
      el("span", { cls: "nm", text: firstName(p.name) }),
    ]);
    a.style.setProperty("--d", (-Math.random() * 2.6).toFixed(2) + "s");   // cada uno camina desfasado
    Interact.bindAvatar(a);
    this.layer.append(a);
    this.avatars.set(p.id, a);
    return a;
  },

  // Sincroniza todos los avatares con el estado que mandó el servidor.
  update(patients, rooms) {
    this.syncRooms(rooms);
    for (const p of patients) {
      const a = this.ensureAvatar(p);
      a.dataset.state = p.state;
      a.dataset.level = p.level;
      a.dataset.zone = this.zoneKeyFor(p);
      a.setAttribute("aria-pressed", String(Store.selectedId === p.id));
      a.setAttribute("aria-label", `${p.name}, nivel ${p.level}, ${STATE_LABEL[p.state].toLowerCase()}. Enter para seleccionar.`);
      a.title = `${p.name} (${p.id})`;
    }
    this.layout();
    if (!this.layer.classList.contains("ready")) {
      // Hasta después del primer dibujo las transiciones están apagadas (CSS .ready).
      requestAnimationFrame(() => requestAnimationFrame(() => this.layer.classList.add("ready")));
    }
  },

  // Mueve un avatar a otra zona ya mismo (sin esperar al servidor) para que "camine".
  moveTo(id, zoneKey) {
    const a = this.avatars.get(id);
    if (!a) return;
    a.dataset.zone = zoneKey;
    this.layout();
  },

  // Calcula la posición de cada avatar dentro de su zona y la aplica con transform.
  // Se usan offsetLeft/offsetTop (no getBoundingClientRect) para que las animaciones
  // de las tarjetas, como el "temblor", no alteren las posiciones calculadas.
  layout() {
    if (!this.root || !this.root.offsetWidth) return;
    const wide = window.matchMedia("(min-width: 700px)").matches;
    const SIZE = 44, STEP_X = 58, STEP_Y = wide ? 66 : 52;

    const groups = new Map();
    for (const a of this.avatars.values()) {
      if (a.classList.contains("dragging")) continue;      // el que se arrastra sigue al dedo
      const key = a.dataset.zone;
      if (!groups.has(key)) groups.set(key, []);
      groups.get(key).push(a);
    }

    for (const [key, list] of groups) {
      list.sort((x, y) => x.dataset.id.localeCompare(y.dataset.id));
      if (key.startsWith("room-")) {
        const room = this.roomEls.get(Number(key.slice(5)));
        if (!room) continue;
        // Dentro de una sala el paciente va sobre la cama (centro de la sala, un poco abajo).
        const cx = room.offsetLeft + room.offsetWidth / 2;
        const cy = room.offsetTop + room.offsetHeight * 0.52;
        list.forEach((a) => this.place(a, cx - SIZE / 2, cy - 28));
        continue;
      }
      const zone = this.zoneEls.get(key);
      if (!zone) continue;
      const cols = Math.max(1, Math.floor((zone.offsetWidth - 16 + (STEP_X - SIZE)) / STEP_X));
      const used = Math.min(cols, list.length);
      const rowWidth = used * STEP_X - (STEP_X - SIZE);
      const startX = zone.offsetLeft + (zone.offsetWidth - rowWidth) / 2;
      const startY = zone.offsetTop + 32;
      list.forEach((a, i) => this.place(a, startX + (i % cols) * STEP_X, startY + Math.floor(i / cols) * STEP_Y));
    }
  },

  place(a, x, y) {
    a._x = x;
    a._y = y;
    a.style.transform = `translate(${x}px, ${y}px)`;
  },

  setMode(mode) {                        // "attack" | "assign" | null: resalta dónde se puede hacer clic
    if (mode) this.root.dataset.mode = mode;
    else delete this.root.dataset.mode;
  },

  markSelected() {
    for (const [id, a] of this.avatars) a.setAttribute("aria-pressed", String(Store.selectedId === id));
  },

  // Número de la sala que está bajo el puntero (o null). Sirve para soltar un paciente.
  roomAt(clientX, clientY) {
    for (const [num, btn] of this.roomEls) {
      const r = btn.getBoundingClientRect();
      if (clientX >= r.left && clientX <= r.right && clientY >= r.top && clientY <= r.bottom) return num;
    }
    return null;
  },

  // Resalta la cama sobre la que se está arrastrando: turquesa si está libre, coral si está llena.
  setDrop(num) {
    for (const btn of this.roomEls.values()) btn.classList.remove("drop-ok", "drop-full");
    if (num == null) return;
    const room = Store.data.rooms.find((r) => r.number === num);
    const free = room && room.state === "Available";
    this.roomEls.get(num).classList.add(free ? "drop-ok" : "drop-full");
  },

  shakeRoom(num) {
    const btn = this.roomEls.get(num);
    if (!btn) return;
    btn.classList.remove("shake");
    void btn.offsetWidth;                // reinicia la animación si ya estaba puesta
    btn.classList.add("shake");
    setTimeout(() => btn.classList.remove("shake"), 600);
  },
};

/* =====================================================================
   5. PANEL — detalle del paciente seleccionado y lista de personal
   ===================================================================== */
const Panel = {
  zoneChoice: "cafeteria",   // zona elegida en el <select> del panel
  signature: "",             // para no redibujar (y no perder el foco) si nada cambió

  render() {
    const p = Store.selectedId ? Store.patient(Store.selectedId) : null;
    const sig = p ? [p.id, p.state, p.room, p.location, p.doctor].join("|") : "none";
    if (sig === this.signature) return;
    this.signature = sig;

    const box = $("#detail-body");
    if (!p) {
      box.replaceChildren(el("div", { cls: "detail-empty" }, [
        icon("moon"),
        el("p", { text: "Ningún paciente seleccionado" }),
        el("p", { text: "Haz clic en un avatar del plano para ver sus datos y las acciones disponibles." }),
      ]));
      return;
    }

    const lvl = p.level.toLowerCase();
    const parts = [
      el("div", { cls: "pt-head" }, [
        el("div", { cls: "pt-avatar lvl-" + lvl, text: initials(p.name) }),
        el("div", {}, [el("div", { cls: "pt-name", text: p.name }), el("div", { cls: "pt-id", text: p.id })]),
      ]),
      el("div", { cls: "pt-badges" }, [
        el("span", { cls: "badge b-lvl-" + lvl, text: p.level }),
        el("span", { cls: "badge " + (STATE_BADGE[p.state] || "b-info"), text: STATE_LABEL[p.state] || p.state }),
      ]),
      el("dl", { cls: "pt-data" }, [
        el("dt", { text: "Edad" }), el("dd", { text: p.age + " años" }),
        el("dt", { text: "Ubicación" }), el("dd", { text: p.location || "—" }),
        el("dt", { text: "Sala" }), el("dd", { text: p.room ? "Sala " + p.room : "Sin sala" }),
        el("dt", { text: "Doctor asignado" }), el("dd", { text: p.doctor || "Sin doctor asignado" }),
      ]),
    ];

    const actions = el("div", { cls: "pt-actions" });
    if (p.state === "Awake") {
      const select = el("select", { attrs: { id: "zone-select" } },
        ZONES.map((z) => el("option", { text: z.name, attrs: { value: z.key } })));
      select.value = this.zoneChoice;
      select.addEventListener("change", () => { this.zoneChoice = select.value; });
      const go = el("button", { cls: "btn danger block", attrs: { type: "button", "aria-label": "Provocar ataque de sueño a " + p.name } },
        [icon("moon"), el("span", { text: "Ataque de sueño" })]);
      go.addEventListener("click", () => Actions.attack(p.id, select.value));
      actions.append(el("label", { cls: "field" }, [el("span", { text: "Zona donde le da el ataque" }), select]), go);
    } else {
      const wake = el("button", { cls: "btn secondary block", attrs: { type: "button", "aria-label": "Despertar a " + p.name } },
        [icon("sun"), el("span", { text: "Despertar" })]);
      wake.addEventListener("click", () => Actions.wake(p.id));
      actions.append(wake);
      if (p.state === "AsleepInHallway") {
        const assign = el("button", { cls: "btn block", attrs: { type: "button", "aria-label": "Asignar sala a " + p.name } },
          [icon("bed"), el("span", { text: "Asignar sala" })]);
        assign.addEventListener("click", () => Actions.assign(p.id, null));
        actions.append(assign);
      }
    }
    const clear = el("button", { cls: "btn secondary block", attrs: { type: "button", "aria-label": "Quitar la selección" } },
      [icon("close"), el("span", { text: "Quitar selección" })]);
    clear.addEventListener("click", () => Interact.select(null));
    actions.append(clear);

    box.replaceChildren(...parts, actions);
  },

  // Personal: el tipo concreto (Doctor / Orderly) lo calculó el backend con un type switch.
  renderStaff(staff) {
    $("#staff").replaceChildren(...staff.map((s) => el("li", {}, [
      el("span", { cls: "who" }, [el("span", { text: s.name }), el("small", { text: s.id })]),
      el("span", { cls: "badge " + (s.type === "Doctor" ? "b-info" : "b-warn"), text: s.type === "Doctor" ? "Doctor" : "Orderly (camillero)" }),
    ])));
  },
};

/* =====================================================================
   6. ACTIONS — lo que el usuario puede hacer. Cada acción llama a un
   endpoint, muestra un aviso y al final recarga el estado (refresh).
   La vista NO decide nada: si el servidor dice 409, se muestra el error.
   ===================================================================== */
const Actions = {
  // Ataque de sueño: primero el avatar "camina" hasta la zona (animación) y
  // después se avisa al hospital, que despacha a un doctor o camillero.
  async attack(id, zoneKey) {
    const zone = zoneByKey(zoneKey);
    const p = Store.patient(id);
    if (!zone || !p || p.state !== "Awake") return false;

    const before = Store.zoneOf.get(id);
    Store.zoneOf.set(id, zoneKey);
    Plan.moveTo(id, zoneKey);
    if (!reducedMotion()) await sleep(750);          // dejar que se vea el recorrido

    let ok = false;
    try {
      const d = await Api.post("/api/attack", { patientId: id, location: zone.name });
      ok = true;
      if (d.warning) Toast.show("warn", `${p.name} se durmió en ${zone.name} y no hay sala libre: queda en el pasillo. (${d.warning})`);
      else Toast.show("ok", `${p.name} se durmió en ${zone.name}. ${d.episode}`);
    } catch (err) {
      Store.zoneOf.set(id, before);                  // no pasó: vuelve a donde estaba
      Toast.show("error", err.message);
    }
    await refresh();
    return ok;
  },

  async wake(id) {
    try {
      const d = await Api.post("/api/wakeup", { patientId: id });
      Toast.show("ok", d.message);
    } catch (err) {
      Toast.show("error", err.message);
    }
    await refresh();
  },

  // Asignar sala. targetRoom es la cama donde el usuario soltó al paciente (o null
  // si se pidió desde un botón). OJO: el hospital siempre da la PRIMERA sala libre;
  // la vista no puede elegir una sala concreta, solo avisar si no coincide.
  async assign(id, targetRoom) {
    const p = Store.patient(id);
    if (!p) return false;
    const rooms = Store.data.rooms;

    if (targetRoom != null) {
      const target = rooms.find((r) => r.number === targetRoom);
      const anyFree = rooms.some((r) => r.state === "Available");
      if (target && target.state !== "Available" && anyFree) {
        // La cama que eligió está llena pero hay otras libres: no se llama al servidor.
        Plan.shakeRoom(targetRoom);
        Toast.show("error", `La sala ${targetRoom} está ocupada. Suelta a ${p.name} sobre una cama libre.`);
        return false;
      }
    }

    let ok = false;
    try {
      const d = await Api.post("/api/assign", { patientId: id });
      ok = true;
      // Confeti sobre la sala cuando el avatar ya llegó a su cama.
      setTimeout(() => Confetti.burstAt(Plan.roomEls.get(d.room)), reducedMotion() ? 0 : 700);
      if (targetRoom != null && d.room !== targetRoom) {
        Toast.show("warn", `El hospital asignó la sala ${d.room} (siempre da la primera disponible).`);
      } else {
        Toast.show("ok", d.message);
      }
    } catch (err) {
      if (targetRoom != null) Plan.shakeRoom(targetRoom);
      Toast.show("error", err.message);
    }
    await refresh();
    return ok;
  },

  async reset() {
    try {
      Carnival.stop();                         // si el carnaval estaba corriendo, se detiene
      await Api.post("/api/reset");
      Store.reset();
      Toast.show("info", "Escenario reiniciado.");
    } catch (err) {
      Toast.show("error", err.message);
    }
    await refresh();
  },
};

/* =====================================================================
   7. INTERACT — clics, doble clic y arrastrar-y-soltar
   - Clic en avatar: selecciona (abre el panel de detalle).
   - Clic en avatar despierto + clic en zona: ataque de sueño allí.
   - Arrastrar un dormido del pasillo hasta una cama libre: asigna sala.
   - Doble clic en un dormido: despertar.
   El arrastre usa "pointer events", que funcionan igual con mouse, dedo
   y lápiz, y también hay alternativas con botones para usar el teclado.
   ===================================================================== */
const Interact = {
  suppressClick: false,    // tras arrastrar, el navegador dispara un "click" que hay que ignorar

  bindAvatar(a) {
    a.addEventListener("click", () => {
      if (this.suppressClick) return;
      this.select(a.dataset.id === Store.selectedId ? null : a.dataset.id);
    });
    a.addEventListener("dblclick", () => {
      const p = Store.patient(a.dataset.id);
      if (p && p.state !== "Awake") Actions.wake(p.id);
    });
    a.addEventListener("pointerdown", (e) => this.dragStart(e, a));
  },

  select(id) {
    Store.selectedId = id;
    this.sync();
  },

  // Deja la selección, el modo del plano, el texto de ayuda y el panel coherentes
  // con el último estado del servidor. Se llama al seleccionar y tras cada refresh.
  sync() {
    const p = Store.selectedId ? Store.patient(Store.selectedId) : null;
    if (!p) Store.selectedId = null;
    Plan.markSelected();
    Plan.setMode(!p ? null : p.state === "Awake" ? "attack" : p.state === "AsleepInHallway" ? "assign" : null);
    Panel.render();
    $("#plan-hint").textContent = this.hint(p);
  },

  hint(p) {
    if (!p) return "Haz clic en un paciente despierto y luego en una zona para provocarle un ataque de sueño. Arrastra a un dormido del pasillo hasta una cama libre.";
    if (p.state === "Awake") return `Elige la zona donde le dará el ataque de sueño a ${p.name}.`;
    if (p.state === "AsleepInHallway") return `Arrastra a ${p.name} hasta una cama libre (o haz clic en una sala). Doble clic para despertar.`;
    return `${p.name} duerme en la sala ${p.room}. Doble clic para despertar.`;
  },

  // Clic en una zona del plano (cafetería, recepción...).
  zoneClick(zoneKey) {
    const p = Store.selectedId ? Store.patient(Store.selectedId) : null;
    if (!p) { Toast.show("info", "Primero selecciona un paciente despierto."); return; }
    if (p.state !== "Awake") { Toast.show("warn", `${p.name} ya está dormido. Solo los pacientes despiertos pueden tener un ataque.`); return; }
    Actions.attack(p.id, zoneKey);
  },

  // Clic en una sala: sirve como alternativa al arrastre (teclado y celular).
  roomClick(roomNumber) {
    const p = Store.selectedId ? Store.patient(Store.selectedId) : null;
    if (!p) { Toast.show("info", "Selecciona un paciente dormido en el pasillo para asignarle sala."); return; }
    if (p.state === "AsleepInHallway") Actions.assign(p.id, roomNumber);
    else if (p.state === "Awake") Toast.show("warn", "Las salas son para pacientes dormidos. Elige una zona para el ataque de sueño.");
    else Toast.show("info", `${p.name} ya tiene sala.`);
  },

  // --- Arrastrar y soltar ---
  dragStart(e, a) {
    const p = Store.patient(a.dataset.id);
    if (!p || p.state !== "AsleepInHallway" || (e.pointerType === "mouse" && e.button !== 0)) return;

    const startX = e.clientX, startY = e.clientY;
    // Distancia entre el dedo y la esquina del avatar, para que no "salte" al agarrarlo.
    const base0 = Plan.root.getBoundingClientRect();
    const grabX = e.clientX - (base0.left + a._x);
    const grabY = e.clientY - (base0.top + a._y);
    let dragging = false;
    try { a.setPointerCapture(e.pointerId); } catch (_) { /* no es grave */ }

    const move = (ev) => {
      // Hasta moverse unos píxeles es un clic normal, no un arrastre.
      if (!dragging && Math.hypot(ev.clientX - startX, ev.clientY - startY) < 6) return;
      if (!dragging) {
        dragging = true;
        a.classList.add("dragging");
        if (Store.selectedId !== p.id) this.select(p.id);
      }
      const base = Plan.root.getBoundingClientRect();
      a.style.transform = `translate(${ev.clientX - base.left - grabX}px, ${ev.clientY - base.top - grabY}px)`;
      Plan.setDrop(Plan.roomAt(ev.clientX, ev.clientY));    // resalta la cama de abajo
    };

    const finish = async (ev, cancelled) => {
      a.removeEventListener("pointermove", move);
      a.removeEventListener("pointerup", up);
      a.removeEventListener("pointercancel", cancel);
      try { a.releasePointerCapture(e.pointerId); } catch (_) { /* ya soltado */ }
      if (!dragging) return;

      this.suppressClick = true;
      setTimeout(() => { this.suppressClick = false; }, 0);
      a.classList.remove("dragging");
      Plan.setDrop(null);
      const room = cancelled ? null : Plan.roomAt(ev.clientX, ev.clientY);
      Plan.layout();                                         // vuelve a su sitio mientras se resuelve
      if (room != null) await Actions.assign(p.id, room);
    };
    const up = (ev) => finish(ev, false);
    const cancel = (ev) => finish(ev, true);

    a.addEventListener("pointermove", move);
    a.addEventListener("pointerup", up);
    a.addEventListener("pointercancel", cancel);
  },
};

// Escape quita la selección.
document.addEventListener("keydown", (e) => {
  if (e.key === "Escape" && Store.selectedId) Interact.select(null);
});

/* =====================================================================
   FX — contadores animados y confeti carnavalero
   Todo respeta prefers-reduced-motion: si el usuario pidió menos
   movimiento, los números cambian de golpe y no hay confeti.
   ===================================================================== */
const Counters = {
  shown: {},                     // clave -> número que se está mostrando

  // Calcula los números a partir de lo que respondió el servidor.
  compute() {
    const { patients, rooms } = Store.data;
    return {
      awake: patients.filter((p) => p.state === "Awake").length,
      hallway: patients.filter((p) => p.state === "AsleepInHallway").length,
      bed: patients.filter((p) => p.state === "AsleepInBed").length,
      freeRooms: rooms.filter((r) => r.state === "Available").length,
      episodes: Counters.episodeCount(),
    };
  },

  // Episodios del día: suma lo que cada doctor reporta en MyEpisodes (5.2).
  episodeCount() {
    return Store.queries.episodesByDoctor.reduce((n, d) => n + d.episodes.length, 0);
  },

  update() {
    const values = this.compute();
    for (const [key, value] of Object.entries(values)) {
      const node = $(`.counter[data-key="${key}"] .c-num`);
      if (!node) continue;
      const from = this.shown[key] === undefined ? value : this.shown[key];
      this.shown[key] = value;
      this.tween(node, from, value);
    }
  },

  // Cuenta de un número a otro en ~450 ms (easeOutCubic) con requestAnimationFrame.
  tween(node, from, to) {
    cancelAnimationFrame(node._raf);
    if (from === to || reducedMotion()) { node.textContent = String(to); return; }
    node.classList.remove("bump");
    void node.offsetWidth;
    node.classList.add("bump");
    const t0 = performance.now(), DURATION = 450;
    const step = (now) => {
      const k = Math.min(1, (now - t0) / DURATION);
      const eased = 1 - Math.pow(1 - k, 3);
      node.textContent = String(Math.round(from + (to - from) * eased));
      if (k < 1) node._raf = requestAnimationFrame(step);
    };
    node._raf = requestAnimationFrame(step);
  },
};

// Mini explosión de confeti en un <canvas> fijo sobre toda la pantalla.
const Confetti = {
  COLORS: ["#ff6b57", "#ffc93c", "#0fa3b1", "#ff4fa3", "#7a5cff", "#2fd1b0", "#ffffff"],
  parts: [],
  raf: 0,

  burst(x, y, amount = 90) {
    if (reducedMotion()) return;
    const canvas = $("#confetti");
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    canvas.width = window.innerWidth * dpr;
    canvas.height = window.innerHeight * dpr;
    this.ctx = canvas.getContext("2d");
    this.ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

    for (let i = 0; i < amount; i++) {
      this.parts.push({
        x, y,
        vx: rand(-7, 7), vy: rand(-13, -3),      // sale disparado hacia arriba y a los lados
        size: rand(5, 10), rot: rand(0, 6.28), vr: rand(-.3, .3),
        color: pick(this.COLORS), round: Math.random() < .3,
        life: rand(70, 110),                      // cuadros que vive (~1.5 s)
      });
    }
    if (!this.raf) this.raf = requestAnimationFrame(() => this.frame());
  },

  frame() {
    const ctx = this.ctx;
    ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);
    this.parts = this.parts.filter((p) => p.life > 0 && p.y < window.innerHeight + 20);
    for (const p of this.parts) {
      p.vy += 0.38;                               // gravedad
      p.vx *= 0.99;                               // un poco de roce con el aire
      p.x += p.vx; p.y += p.vy; p.rot += p.vr; p.life--;
      ctx.save();
      ctx.translate(p.x, p.y);
      ctx.rotate(p.rot);
      ctx.globalAlpha = Math.min(1, p.life / 25);
      ctx.fillStyle = p.color;
      if (p.round) { ctx.beginPath(); ctx.arc(0, 0, p.size / 2, 0, 6.28); ctx.fill(); }
      else ctx.fillRect(-p.size / 2, -p.size / 3, p.size, p.size * 0.66);
      ctx.restore();
    }
    if (this.parts.length) this.raf = requestAnimationFrame(() => this.frame());
    else { this.raf = 0; ctx.clearRect(0, 0, window.innerWidth, window.innerHeight); }
  },

  // Confeti saliendo del centro de un elemento (por ejemplo la sala recién asignada).
  burstAt(element) {
    if (!element) return;
    const r = element.getBoundingClientRect();
    this.burst(r.left + r.width / 2, r.top + r.height / 2);
  },
};

/* =====================================================================
   MODO CARNAVAL — simulación hecha en el navegador
   Cada 2 a 3 segundos elige al azar un paciente despierto y una zona, y llama
   a Actions.attack (o sea, al mismo POST /api/attack que usa un clic normal).
   No hay un "bucle" ni setInterval: cada vuelta (tick) espera a que termine el
   ataque anterior y recién entonces programa la siguiente con setTimeout.
   Así nunca se cruzan dos ataques y el cambio de velocidad se nota de inmediato.
   Se detiene solo cuando ya nadie está despierto.
   ===================================================================== */
const Carnival = {
  running: false,
  timer: 0,
  speed: 3,                                   // 1 (muy lenta) ... 5 (muy rápida)
  LABELS: ["Muy lenta", "Lenta", "Normal", "Rápida", "Muy rápida"],
  FACTOR: [1.7, 1.3, 1, 0.7, 0.45],           // multiplica la espera base de 2 a 3 s

  init() {
    $("#carnival-btn").addEventListener("click", () => (this.running ? this.stop() : this.start()));
    const slider = $("#carnival-speed");
    slider.addEventListener("input", () => this.setSpeed(Number(slider.value)));
    this.setSpeed(this.speed);
  },

  setSpeed(n) {
    this.speed = n;
    const slider = $("#carnival-speed");
    const label = this.LABELS[n - 1];
    slider.style.setProperty("--fill", ((n - 1) / 4) * 100 + "%");   // relleno de la barra
    slider.setAttribute("aria-valuetext", label);
    $("#carnival-speed-label").textContent = label;
  },

  waitMs() { return rand(2000, 3000) * this.FACTOR[this.speed - 1]; },

  start() {
    if (!Store.awake().length) { Toast.show("info", "Nadie está despierto: despierta a alguien para empezar el carnaval."); return; }
    this.running = true;
    this.paint();
    Toast.show("info", "¡Arranca el carnaval! Pausa cuando quieras.");
    this.schedule();
  },

  // reason: texto opcional para avisar por qué se detuvo
  stop(reason) {
    clearTimeout(this.timer);
    const was = this.running;
    this.running = false;
    this.paint();
    if (was && reason) Toast.show("info", reason);
  },

  schedule() {
    clearTimeout(this.timer);
    this.timer = setTimeout(() => this.tick(), this.waitMs());
  },

  async tick() {
    if (!this.running) return;
    const awake = Store.awake();
    if (!awake.length) { this.stop("Todos los pacientes están dormidos: el carnaval terminó solo."); return; }
    await Actions.attack(pick(awake).id, pick(ZONES).key);          // paciente y zona al azar
    if (!this.running) return;                                      // lo pausaron mientras tanto
    if (!Store.awake().length) { this.stop("Todos los pacientes están dormidos: el carnaval terminó solo."); return; }
    this.schedule();
  },

  // Botón: texto, ícono y aria-pressed según esté corriendo o en pausa.
  paint() {
    const btn = $("#carnival-btn");
    btn.setAttribute("aria-pressed", String(this.running));
    btn.classList.toggle("running", this.running);
    btn.setAttribute("aria-label", this.running ? "Pausar el modo carnaval" : "Iniciar el modo carnaval: ataques de sueño aleatorios");
    btn.replaceChildren(icon(this.running ? "pause" : "play"), el("span", { text: this.running ? "Pausar carnaval" : "Modo carnaval" }));
  },
};

/* =====================================================================
   9. ARRANQUE — refresh() es el único sitio que lee del servidor
   ===================================================================== */
async function refresh() {
  try {
    const [state, queries] = await Promise.all([Api.get("/api/state"), Api.get("/api/queries")]);
    Store.data = state;
    Store.queries = queries;
    Store.seedZones(state.patients);

    document.title = state.hospital;
    Plan.update(state.patients, state.rooms);
    Panel.renderStaff(state.staff);
    Interact.sync();
    Counters.update();
  } catch (err) {
    Toast.show("error", "No se pudo cargar el estado: " + err.message);
  }
}

Plan.init();
Carnival.init();
$("#reset-btn").addEventListener("click", () => Actions.reset());
refresh();
