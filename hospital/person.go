package hospital

type Person struct {
	id   string
	name string
	age  int
}

func newPerson(id, name string, age int) Person {
	return Person{id: id, name: name, age: age}
}

func NewPerson(id, name string, age int) Person {
	return newPerson(id, name, age)
}

func (p Person) ID() string   { return p.id }
func (p Person) Name() string { return p.name }
func (p Person) Age() int     { return p.age }
