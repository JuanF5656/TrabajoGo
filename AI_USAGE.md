# Uso de IA en el proyecto


## Herramientas usadas y para qué

- **Claude Code (CLI)** — sesión del 17/10/2026 para:
- **Claude Code — interfaz web (bonus).** Se usó para escribir `cmd/web` (servidor
  `net/http`, DTOs JSON, página con HTML/CSS/JS, tests con `httptest`). La instrucción
  fue explícita: no modificar nada dentro de `hospital/` y resolver las carencias en la
  capa web. Lo que faltaba (el paquete no expone la lista de doctores) se resolvió
  guardando las referencias en el servidor web. Se verificó con `git diff --stat HEAD --
  hospital/`, que no mostró cambios. Se probó a mano con `curl` (4 ataques, 409 sin sala,
  despertar y reasignar) antes de escribir los tests. Límite: en la máquina de trabajo no
  hay compilador C, así que `-race` no se pudo ejecutar ahí; hay que correrlo en otro
  equipo y no se debe afirmar que pasó sin haberlo hecho.



## Qué aprendió cada integrante

- **Thomas:** Me apoye con claude code para apoyarme en GO, para crear algunos metodos con los que no estaba familiarizado, tambien nos ayudo a corregir bugs y errores.
Un bug que nos corrigio fue: - Al renombrar un error centinela (`ErrPatientAlreadyAdmited` → `ErrPatientAlreadyAdmitted`), el bloque `var (...)` quedó desalineado y `gofmt -l .` lo marcó. Lo corregimos con `gofmt -w .`. Desde entonces corremos `gofmt` y `go vet` antes de cada commit.
- **Juan Felipe:** La IA me ayudó a tener más claro la sintaxis de go y darme cuenta de errores lógicos que estaba cometiendo
BUG y corregido por la IA
episode := EpisodeRecord{
    id:         fmt.Sprintf("EP-%03d", len(d.hospital.history)+1),
    at:         time.Now(),
    patient:    p,
    attendedBy: d,
    location:   location,
    room:       room,
}
aqui me ayudo a darme cuenta que habia que cambiar el nombre de los campos por que en episode record y en doctor tenian nombres diferentes y no me habia dado cuenta
"ocurredAt" lo cambié a "at" y "attender" por "attendedBy"