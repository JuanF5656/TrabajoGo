# Hospital de los Costeños con Narcolepsia

## Integrantes

- Persona A: Thomas Gonzales
- Persona B: Juan Felipe Atehortua

## Descripción

Sistema de gestión hospitalaria desarrollado en Go para representar pacientes con narcolepsia, doctores, camilleros, habitaciones y episodios de sueño repentino.

## Requisitos

- Go 1.21 o superior.

## Ejecución

Desde la raíz del repositorio:

```bash
go run .
```

## Pruebas

```bash
go test ./...
```

## Verificaciones de código

```bash
gofmt -w .
go vet ./...
```

## Decisiones de diseño

- **Composición:** `Patient` y `Doctor` reutilizan la estructura `Person` mediante embedding.
- **Polimorfismo:** el hospital almacena al personal mediante la interfaz `Attender`.
- **Encapsulación:** los campos internos de las entidades no se exportan directamente.
- **Manejo de errores:** las operaciones que pueden fallar devuelven errores en lugar de utilizar `panic`.
- **Separación de responsabilidades:** el paquete `hospital` contiene la lógica del dominio y `main.go` construye el escenario e imprime los resultados.

## Escenario de demostración

Se crean dos doctores, un camillero, tres habitaciones y cinco pacientes. Se simulan cuatro ataques, incluyendo un caso en el que no hay camas disponibles. Posteriormente se libera y reasigna una habitación y se imprimen las cuatro consultas requeridas.
