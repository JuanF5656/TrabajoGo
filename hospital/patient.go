package hospital

import (
	"fmt"
	"strings"
)

// Patient es un costeño con narcolepsia admitido en el hospital.
// Embebe Person: p.ID() y p.Name() vienen de Person sin escribirlos de nuevo.
type Patient struct {
	Person
	level           NarcolepsyLevel
	state           PatientState
	currentLocation string
	assignedDoctor  *Doctor
	room            *Room // sala donde duerme; nil si está despierto o en el pasillo
}

// NewPatient valida los datos y crea un paciente despierto.
// Devuelve error en vez de hacer panic: los datos malos son previsibles.
func NewPatient(id, name string, age int, level NarcolepsyLevel) (*Patient, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("%w: id and name are required", ErrInvalidPatient)
	}
	if age < 0 {
		return nil, fmt.Errorf("%w: age %d is negative", ErrInvalidPatient, age)
	}
	if !level.isValid() {
		return nil, fmt.Errorf("%w: unknown narcolepsy level %d", ErrInvalidPatient, int(level))
	}
	return &Patient{
		Person: newPerson(id, name, age),
		level:  level,
		state:  Awake,
	}, nil
}

// --- Getters (los campos son privados) ---

func (p *Patient) Level() NarcolepsyLevel  { return p.level }
func (p *Patient) State() PatientState     { return p.state }
func (p *Patient) Location() string        { return p.currentLocation }
func (p *Patient) AssignedDoctor() *Doctor { return p.assignedDoctor }
func (p *Patient) Room() *Room             { return p.room }

// --- Comportamiento ---

// SufferSleepAttack: el paciente se queda dormido de repente en location.
// Queda en AsleepInHallway hasta que el hospital le asigne una sala.
// Receiver puntero: modifica el estado del paciente original, no una copia.
func (p *Patient) SufferSleepAttack(location string) error {
	if p.state != Awake {
		return fmt.Errorf("%w: %s (%s)", ErrAlreadyAsleep, p.id, p.state)
	}
	if strings.TrimSpace(location) == "" {
		location = "unknown location"
	}
	p.state = AsleepInHallway
	p.currentLocation = location
	return nil
}

// WakeUp despierta al paciente. Si estaba en una cama, libera la sala
// para que otro paciente la pueda usar.
func (p *Patient) WakeUp() error {
	if p.state == Awake {
		return fmt.Errorf("%w: %s", ErrAlreadyAwake, p.id)
	}
	if p.room != nil {
		if err := p.room.Release(p); err != nil {
			return fmt.Errorf("waking up %s: %w", p.id, err)
		}
		p.room = nil
	}
	p.state = Awake
	return nil
}

// moveToBed lo usa el hospital (AssignRoom) después de ocupar la sala.
// Es privado: desde main no se puede poner a alguien "en cama" sin sala.
func (p *Patient) moveToBed(r *Room) {
	p.room = r
	p.state = AsleepInBed
	p.currentLocation = fmt.Sprintf("room %d", r.Number())
}

// assignDoctor lo usa Doctor.DiagnosePatient (Persona B).
func (p *Patient) assignDoctor(d *Doctor) {
	p.assignedDoctor = d
}

// String permite imprimir un paciente con fmt (implementa fmt.Stringer).
func (p *Patient) String() string {
	return fmt.Sprintf("%s %s (%s, %s)", p.id, p.name, p.level, p.state)
}
