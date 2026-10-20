package hospital

import (
	"fmt"
<<<<<<< HEAD
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
=======
	"time"
)

// Orderly representa a un camillero del hospital.
type Orderly struct {
	Person
	hospital *Hospital
	episodes []EpisodeRecord
}

func (o *Orderly) Available() bool {
	return o != nil && o.hospital != nil
}

// NewOrderly crea un camillero asociado al hospital.
func NewOrderly(
	id string,
	name string,
	age int,
	hospital *Hospital,
) *Orderly {
	return &Orderly{
		Person:   newPerson(id, name, age),
		hospital: hospital,
	}
}

// MyEpisodes devuelve una copia del historial del camillero.
func (o *Orderly) MyEpisodes() []EpisodeRecord {
	result := make([]EpisodeRecord, len(o.episodes))
	copy(result, o.episodes)
	return result
}

// Attend atiende a un paciente y trata de asignarle una habitación.
func (o *Orderly) Attend(
	p *Patient,
	location string,
) (EpisodeRecord, error) {
	if o == nil || o.hospital == nil {
		return EpisodeRecord{}, fmt.Errorf(
			"el camillero no está asociado a un hospital",
		)
	}

	if p == nil {
		return EpisodeRecord{}, ErrNilPatient
	}

	if !o.hospital.isAdmitted(p) {
		return EpisodeRecord{}, fmt.Errorf(
			"%w: %s", ErrPatientNotAdmitted, p.ID(),
		)
	}

	// Si está despierto, comienza el episodio.
	// Si ya está dormido en el pasillo, continúa la atención.
	if p.State() != AsleepInHallway {
		if err := p.SufferSleepAttack(location); err != nil {
			return EpisodeRecord{}, err
		}
	}

	// El hospital se encarga de encontrar y ocupar una habitación.
	room, roomErr := o.hospital.AssignRoom(p)

	// Conservamos el episodio aunque no haya habitaciones disponibles.
	episode := EpisodeRecord{
		id:         fmt.Sprintf("EP-%03d", len(o.hospital.history)+1),
		at:         time.Now(),
		patient:    p,
		attendedBy: o,
		location:   location,
		room:       room,
	}

	if err := o.hospital.RegisterEpisode(episode); err != nil {
		return EpisodeRecord{}, err
	}

	o.episodes = append(o.episodes, episode)

	return episode, roomErr
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
}
