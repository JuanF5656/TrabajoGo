package hospital

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// Orderly es un camillero: lleva al paciente dormido a una sala. A propósito
// NO tiene DiagnosePatient: solo los doctores pueden diagnosticar.
type Orderly struct {
	Person
	hospital  *Hospital   // se asigna al contratarlo con Hospital.Hire
	available atomic.Bool // atómico: se puede cambiar mientras el hospital despacha
}

// NewOrderly valida los datos y crea un camillero disponible.
func NewOrderly(id, name string, age int) (*Orderly, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("%w: id and name are required", ErrInvalidStaff)
	}
	if age < 0 {
		return nil, fmt.Errorf("%w: age %d is negative", ErrInvalidStaff, age)
	}
	o := &Orderly{Person: newPerson(id, name, age)}
	o.available.Store(true)
	return o, nil
}

// Available indica si el camillero puede atender un nuevo episodio.
func (o *Orderly) Available() bool { return o.available.Load() }

// SetAvailable marca al camillero como disponible o no.
func (o *Orderly) SetAvailable(v bool) { o.available.Store(v) }

// setHospital lo llama Hospital.Hire al contratar al camillero.
func (o *Orderly) setHospital(h *Hospital) { o.hospital = h }

// Attend atiende un ataque de sueño igual que el doctor: usa la misma lógica
// del hospital, así no se duplica código entre Doctor y Orderly.
func (o *Orderly) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if o.hospital == nil {
		return EpisodeRecord{}, fmt.Errorf("%w: orderly %s", ErrNotHired, o.ID())
	}
	return o.hospital.handleSleepAttack(o, p, location)
}
