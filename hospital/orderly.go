package hospital

import (
	"fmt"
	"time"
)

// Orderly representa a un camillero del hospital.
type Orderly struct {
	Person
	hospital *Hospital
	episodes []EpisodeRecord
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

	o.hospital.history = append(o.hospital.history, episode)
	o.episodes = append(o.episodes, episode)

	return episode, roomErr
}
