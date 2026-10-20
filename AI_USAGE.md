# Uso de IA en el proyecto

Este documento deja constancia de cómo se usó la IA en el taller. Lo que no se pudo
verificar desde el historial de git o desde la sesión de trabajo está marcado con
`[COMPLETAR]` para que lo llenen los integrantes.

## Herramientas usadas y para qué

- **Claude Code (CLI)** — sesión del 19/10/2026 para:
  - sacar del repositorio el submódulo roto `TrabajoGo`;
  - implementar el despacho polimórfico (`DispatchSleepAttack`), `Orderly`, `Doctor`
    completo y `EpisodeRecord`;
  - reescribir `main.go` para que siga la sección 6;
  - restaurar comentarios, escribir tests nuevos, el README y este documento.
- **Claude Code — interfaz web (bonus).** Se usó para escribir `cmd/web` (servidor
  `net/http`, DTOs JSON, página con HTML/CSS/JS, tests con `httptest`). La instrucción
  fue explícita: no modificar nada dentro de `hospital/` y resolver las carencias en la
  capa web. Lo que faltaba (el paquete no expone la lista de doctores) se resolvió
  guardando las referencias en el servidor web. Se verificó con `git diff --stat HEAD --
  hospital/`, que no mostró cambios. Se probó a mano con `curl` (4 ataques, 409 sin sala,
  despertar y reasignar) antes de escribir los tests. Límite: en la máquina de trabajo no
  hay compilador C, así que `-race` no se pudo ejecutar ahí; hay que correrlo en otro
  equipo y no se debe afirmar que pasó sin haberlo hecho.
- [COMPLETAR: otras herramientas de IA que usaron antes (por ejemplo, para el commit
  "Describe lo que cambiaste" o para la parte de Persona A), y para qué.]

## 3 prompts decisivos

1. **El prompt de corrección final (esta sesión).** Pidió, en orden y con commits
   pequeños: quitar el submódulo, despacho polimórfico con `[]Attender`, completar
   `Doctor`, ajustar `main.go` a la sección 6, restaurar comentarios, tests,
   README y este archivo. Fue decisivo porque fijaba el orden, las reglas
   (`gofmt`, `go vet`, `go test` tras cada punto) y no cambiar firmas existentes.
2. [COMPLETAR: prompt decisivo 2 — copiar el texto real que usaron.]
3. [COMPLETAR: prompt decisivo 3 — copiar el texto real que usaron.]

## Casos donde la IA se equivocó y cómo se corrigió

1. **Hospital sin despacho polimórfico.** El enunciado exige que el hospital despache
   a quien esté disponible sin conocer su tipo. En el código heredado `Hospital` tenía
   el campo `staff []Attender`, pero nada lo usaba, y `Doctor` era solo un struct con
   `Person` (sin `Attend`, sin `Available`). Corrección: se agregó `Available()` a la
   interfaz `Attender`, `Hospital.Hire` y `DispatchSleepAttack` con round-robin; `main`
   dejó de armar su propia lista de attenders.
2. **Reasignación con un segundo `Attend`.** Una forma natural de "darle sala" al
   paciente que quedó en el pasillo era volver a llamar a `Attend`. Eso crea un
   episodio falso y hace que `SevereReport` cuente 2 episodios para quien tuvo 1.
   Corrección: se usa `h.AssignRoom(p)` directamente, y hay un test
   (`TestSevereReport_FullScenarioReassignedCountsOnce`) que lo vigila.
3. **Salida no determinista por el orden de un mapa.** `SevereReport` devuelve un
   `map[*Patient]int`; en Go el orden de iteración de un mapa es aleatorio, así que
   imprimirlo directo cambiaba la salida en cada ejecución. Corrección: `main` copia
   los pacientes a un slice y los ordena por ID con `sort.Slice` antes de imprimir.
4. **Comentarios de documentación borrados.** El commit `a0d9f57` ("varios Cambios")
   eliminó los comentarios de doc de `states.go`, `person.go`, `room.go`,
   `hospital.go`, `errors.go` y `hospital_test.go`, y dejó los archivos sin pasar por
   `gofmt` (líneas en blanco dobles). Corrección: se recuperaron desde `ac5a07e` sin
   revertir el código y se aplicó `gofmt`.
5. **Submódulo roto.** El repositorio tenía `TrabajoGo` registrado como submódulo
   (gitlink `160000`) sin archivo `.gitmodules`, así que no se podía clonar bien.
   Corrección: `git rm --cached TrabajoGo` y la carpeta se movió fuera del repo.
6. **Piezas que se asumían hechas y no existían.** El encargo suponía que existían
   `NewDoctor`, `Orderly`, `RegisterEpisode` y un README; en el repo no estaban
   (`main.go` estaba vacío y `Doctor` sin comportamiento). Se implementaron y se
   avisó en vez de asumir que ya estaban.

## Qué aprendió cada integrante

- **Thomas:** [COMPLETAR: Thomas]
- **Juan Felipe:** [COMPLETAR: Juan Felipe]
