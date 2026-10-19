package hospital

// Person agrupa los datos de identidad que comparten Patient y Doctor.
// Se EMBEBE (composición) en ambos en lugar de copiar los tres campos:
// así Patient y Doctor "heredan" los métodos ID(), Name() y Age().
//
// Los campos van en minúscula (no exportados): desde fuera del paquete
// hospital solo se pueden leer mediante los getters, nunca modificar.
type Person struct {
	id   string
	name string
	age  int
}

// newPerson es interno: solo los constructores del paquete (NewPatient,
// NewDoctor) crean personas, y ellos validan los datos.
func newPerson(id, name string, age int) Person {
	return Person{id: id, name: name, age: age}
}

// Receivers por valor: Person es pequeño y estos métodos solo leen.
// Así funcionan igual sobre Person, *Person, Patient y *Patient.

func (p Person) ID() string   { return p.id }
func (p Person) Name() string { return p.name }
func (p Person) Age() int     { return p.age }
