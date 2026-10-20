package hospital

import (
	"fmt"
	"time"
)

// Attender: cualquier miembro del personal que puede atender un episodio.
type Attender interface {
	ID() string
	Name() string
	Available() bool // añadido: "despachar a quien esté disponible"
	Attend(p *Patient, location string) (EpisodeRecord, error)
}

func NewDoctor(person Person, hospital *Hospital) *Doctor {
	return &Doctor{
		Person:   person,
		hospital: hospital,
	}
}

type Doctor struct {
	Person
	patients []*Patient
	episodes []EpisodeRecord
	hospital *Hospital
}

func (d *Doctor) Available() bool {
	return d != nil && d.hospital != nil
}

func (d *Doctor) MyEpisodes() []EpisodeRecord {
	result := make([]EpisodeRecord, len(d.episodes))
	copy(result, d.episodes)
	return result
}

func (d *Doctor) DiagnosePatient(p *Patient) error {
	if d == nil || d.hospital == nil {
		return fmt.Errorf("el doctor no está asociado a un hospital")
	}

	if p == nil {
		return ErrNilPatient
	}

	if !d.hospital.isAdmitted(p) {
		return fmt.Errorf("%w: %s", ErrPatientNotAdmitted, p.ID())
	}

	// Evita registrar dos veces al mismo paciente.
	for _, existing := range d.patients {
		if existing == p {
			p.assignDoctor(d)
			return nil
		}
	}

	p.assignDoctor(d)
	d.patients = append(d.patients, p)

	return nil
}

func (d *Doctor) Attend(
	p *Patient,
	location string,
) (EpisodeRecord, error) {
	if d == nil || d.hospital == nil {
		return EpisodeRecord{}, fmt.Errorf(
			"el doctor no está asociado a un hospital",
		)
	}

	if p == nil {
		return EpisodeRecord{}, ErrNilPatient
	}

	if !d.hospital.isAdmitted(p) {
		return EpisodeRecord{}, fmt.Errorf(
			"%w: %s", ErrPatientNotAdmitted, p.ID(),
		)
	}

	// Si el paciente está despierto, registra el inicio del episodio.
	// Si ya está dormido en el pasillo, continúa con la asignación.
	if p.State() != AsleepInHallway {
		if err := p.SufferSleepAttack(location); err != nil {
			return EpisodeRecord{}, err
		}
	}

	// La habitación será nil si no hay ninguna disponible.
	room, roomErr := d.hospital.AssignRoom(p)

	// Construye el registro incluso cuando no se consigue habitación.
	episode := EpisodeRecord{
		id:         fmt.Sprintf("EP-%03d", len(d.hospital.history)+1),
		at:         time.Now(),
		patient:    p,
		attendedBy: d,
		location:   location,
		room:       room,
	}

	// Registra el episodio en el historial general del hospital
	// y en el historial particular del doctor.
	if err := d.hospital.RegisterEpisode(episode); err != nil {
		return EpisodeRecord{}, err
	}

	d.episodes = append(d.episodes, episode)

	return episode, roomErr
}
