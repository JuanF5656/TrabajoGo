# Hospital de los Costeños con Narcolepsia

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
```

## Decisiones de diseño

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


