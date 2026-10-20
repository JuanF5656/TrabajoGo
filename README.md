# Hospital de los Costeños con Narcolepsia

Taller de Programación Orientada a Objetos — Universidad EIA. Escrito en Go 1.21
usando solo la librería estándar.

## Integrantes

- Thomas Andres Gonzalez Aguilar
- Juan Felipe [COMPLETAR: apellido]

## Cómo ejecutarlo

```
go run .          # corre el escenario de la sección 6
go run . -sim     # bonus: simulación concurrente (salida distinta en cada corrida)
go test ./...     # corre los tests del paquete hospital
```

## Estructura

```
main.go                  Arma el escenario e imprime. No tiene lógica de negocio.
hospital/
  states.go              Estados tipados (PatientState, NarcolepsyLevel, RoomState)
  errors.go              Errores centinela (se comparan con errors.Is)
  person.go              Person: datos comunes, embebido en Patient, Doctor y Orderly
  patient.go             Patient y sus transiciones (ataque de sueño, despertar)
  room.go                Room: sala con capacidad
  doctor.go              Interfaz Attender y Doctor
  orderly.go             Orderly (camillero)
  episode.go             EpisodeRecord: registro inmutable de un episodio
  hospital.go            Hospital: personal, despacho, consultas 5.1 a 5.4
  hospital_test.go       Tests
```

## Decisiones de diseño

- **Composición con `Person`.** Patient, Doctor y Orderly embeben `Person` en vez de
  repetir id, nombre y edad. Los campos son privados y se leen con getters.
- **Despacho polimórfico con `DispatchSleepAttack`.** El hospital guarda su personal
  como `[]Attender` y, cuando hay un ataque, recorre la lista y usa al primero con
  `Available() == true`, sin preguntar si es Doctor u Orderly. Reparte en
  *round-robin* (índice privado `nextStaff`) para que no atienda siempre el mismo.
  Si nadie está disponible devuelve `ErrNoStaffAvailable`. Agregar un nuevo tipo de
  personal solo exige implementar `Attender`; el hospital no cambia.
- **Un solo dueño de la lógica del episodio.** `Doctor.Attend` y `Orderly.Attend`
  delegan en `Hospital.handleSleepAttack`, y el id (`EP-001`, `EP-002`...) lo genera
  `Hospital.nextEpisodeID`. Así no se duplica código ni se numera en dos sitios.
- **Solo el doctor diagnostica.** `DiagnosePatient` existe en `Doctor` y no en
  `Orderly`: el compilador impide que un camillero diagnostique.
- **Errores como valores.** Se usan errores centinela envueltos con `%w`, no `panic`.
  Si no hay sala, el episodio igual se registra y se devuelve un error que envuelve
  `ErrNoRoomAvailable`; el paciente queda en el pasillo.
- **Reasignar sala no crea episodios.** Cuando alguien se despierta, el paciente del
  pasillo recibe sala con `AssignRoom` directo. Un segundo `Attend` crearía un
  episodio falso y alteraría el `SevereReport`.
- **Concurrencia (bonus).** `Hospital` tiene un `sync.Mutex`. Los métodos públicos lo
  toman y los internos (`assignRoom`, `registerEpisode`) no, porque el mutex no es
  reentrante y se bloquearía a sí mismo. `simulation.go` lanza un paciente por goroutine.
- **Encapsulamiento.** `Patients()`, `Rooms()` y `Occupants()` devuelven copias.

## Por qué un mapa en SevereReport

`SevereReport` responde "dado un paciente Severe, ¿cuántos episodios tuvo?". Eso es
una búsqueda por clave, que es lo que hace un mapa (`map[*Patient]int`):

1. Se crea una entrada en 0 por cada paciente Severe (así los que no tuvieron
   episodios también aparecen).
2. Se recorre el historial una sola vez y, por cada episodio, se suma 1 si el
   paciente está en el mapa. Es O(n) en una pasada; con un slice habría que buscar al
   paciente en cada episodio.

Un mapa no tiene orden, por eso `main` ordena a los pacientes por ID antes de
imprimir y la salida es siempre la misma.
