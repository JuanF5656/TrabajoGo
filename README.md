# Hospital de los Costeños con Narcolepsia

<<<<<<< HEAD
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
=======
## Integrantes

- **Persona A:** Thomas Gonzales
- **Persona B:** Juan Felipe Atehortua

## Descripción

Sistema de gestión hospitalaria desarrollado en Go que simula la atención de pacientes con narcolepsia. Permite gestionar pacientes, doctores, camilleros, habitaciones y episodios de atención, incluyendo situaciones en las que no hay habitaciones disponibles.

El proyecto aplica conceptos de programación orientada a objetos mediante composición, interfaces, encapsulación y manejo de errores.

## Requisitos

- Go 1.21 o superior.

## Ejecución

Desde la raíz del repositorio, ejecutar:

```bash
go run .
```

El programa construye un escenario de demostración e imprime los resultados de las operaciones realizadas.

## Pruebas

Para ejecutar las pruebas de todos los paquetes:

```bash
go test ./...
```

## Verificaciones de código

Para aplicar el formato estándar de Go:

```bash
gofmt -w .
```

Para analizar posibles errores y problemas comunes:

```bash
go vet ./...
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
```

## Decisiones de diseño

<<<<<<< HEAD
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

## Interfaz web (bonus)

Una interfaz gráfica en el navegador, sin dependencias externas ni internet
(`net/http`, `encoding/json`, `embed`, `sync`).

```
go run ./cmd/web              # y abrir http://localhost:8080
go run ./cmd/web -addr :9000  # otro puerto
```

`go run .` sigue imprimiendo el escenario por consola; el `main.go` de la raíz no cambió.

### Endpoints

| Método | Ruta | Qué hace |
|--------|------|----------|
| GET  | `/api/state` | Hospital, pacientes, salas y personal (Doctor u Orderly por type switch) |
| GET  | `/api/queries` | Consultas 5.1 (pasillo), 5.2 (episodios por doctor), 5.3 (última asignación) y 5.4 (SevereReport ordenado por ID) |
| POST | `/api/attack` `{"patientId","location"}` | `DispatchSleepAttack` (despacho polimórfico) |
| POST | `/api/wakeup` `{"patientId"}` | Despierta al paciente |
| POST | `/api/assign` `{"patientId"}` | `AssignRoom`; 409 si no hay sala |
| POST | `/api/reset` | Reconstruye el escenario inicial |

Los errores salen como `{"error": "..."}`: 400 datos inválidos, 404 paciente inexistente,
409 conflictos (no hay sala, ya dormido, ya despierto). El código se elige con
`errors.Is` sobre los errores centinela del paquete `hospital`. Un ataque sin sala
responde 200 con un `warning`, porque el episodio sí quedó registrado.

### Por qué el paquete hospital no cambió

La regla era separar lógica y vista. El paquete `hospital` no importa `net/http` ni
`encoding/json` y no sabe que existe una web; la web solo llama a su API pública.

- Los objetos del dominio nunca se serializan directo: `cmd/web/api.go` los convierte a
  DTOs con tags `json`, así cambiar la pantalla no obliga a tocar el dominio.
- Lo que el paquete no expone (la lista de doctores para `MyEpisodes`) lo guarda la capa
  web cuando crea el escenario.
- Los handlers HTTP corren en goroutines distintas, por eso todo acceso al hospital va
  dentro de un `sync.Mutex` del servidor web (el mutex interno del hospital no cubre
  secuencias de varias llamadas).
- Se comprueba con `git diff --stat HEAD -- hospital/`, que no muestra cambios.
=======
- **Composición mediante embedding:** `Patient`, `Doctor` y `Orderly` reutilizan la estructura `Person`, heredando su información y sus métodos públicos sin duplicar código.
- **Polimorfismo mediante interfaces:** la interfaz `Attender` define las operaciones comunes del personal encargado de atender pacientes. El hospital puede almacenar y gestionar doctores y camilleros a través de una misma interfaz.
- **Encapsulación:** los campos internos de las entidades no se exportan directamente. El acceso a la información y la modificación del estado se realizan mediante métodos.
- **Manejo de errores:** las operaciones que pueden fallar devuelven errores que permiten identificar situaciones como pacientes inválidos o falta de habitaciones disponibles.
- **Separación de responsabilidades:** el paquete `hospital` contiene las entidades y la lógica del dominio, mientras que `main.go` construye el escenario de demostración y presenta los resultados.
- **Registro de episodios:** cada atención genera un `EpisodeRecord` con información del paciente, el personal responsable, la ubicación, la fecha y la habitación asignada, cuando existe.
- **Gestión de habitaciones:** el hospital asigna habitaciones disponibles y permite liberarlas cuando corresponde. Si no hay disponibilidad, el paciente puede permanecer en el pasillo hasta que se le pueda asignar una habitación.

## Escenario de demostración

El programa realiza las siguientes operaciones:

1. Crea un hospital y registra dos doctores y un camillero.
2. Registra tres habitaciones y admite cinco pacientes.
3. Simula cuatro atenciones en diferentes ubicaciones.
4. Demuestra qué ocurre cuando no hay habitaciones disponibles: el paciente permanece en el pasillo y la atención queda registrada.
5. Despierta a un paciente y libera su habitación.
6. Atiende nuevamente al paciente que estaba en el pasillo y le asigna la habitación liberada.
7. Consulta los pacientes que permanecen en el pasillo, el estado de las habitaciones, el informe de casos severos y los episodios registrados por cada integrante del personal.

## Estructura general del proyecto

- **`main.go`:** punto de entrada del programa y escenario de demostración.
- **`hospital/person.go`:** información y métodos comunes de las personas.
- **`hospital/patient.go`:** estado, ubicación, nivel de narcolepsia y comportamiento de los pacientes.
- **`hospital/doctor.go`:** atención médica y gestión de pacientes asignados.
- **`hospital/orderly.go`:** atención realizada por el camillero.
- **`hospital/room.go`:** creación, ocupación y liberación de habitaciones.
- **`hospital/episode_record.go`:** registro y presentación de los episodios de atención.
- **`hospital/hospital.go`:** administración del personal, los pacientes, las habitaciones y el historial del hospital.


>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
