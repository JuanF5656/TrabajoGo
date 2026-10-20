package hospital

import (
	"fmt"
<<<<<<< HEAD
	"strings"
	"sync/atomic"
=======
	"time"
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
)

// Attender es cualquier miembro del personal que puede atender un episodio.
// El hospital trabaja solo con esta interfaz: no necesita saber si quien
// atiende es un Doctor o un Orderly (polimorfismo).
type Attender interface {
	ID() string
	Name() string
<<<<<<< HEAD
	// Available dice si esta persona puede atender un episodio ahora mismo.
	// Está en la interfaz para que el hospital elija a quién despachar sin
	// preguntar por el tipo concreto (nada de "if doctor ... else if orderly").
	Available() bool
	Attend(p *Patient, location string) (EpisodeRecord, error)
}

// Doctor es un médico del hospital. Embebe Person (id, nombre, edad).
type Doctor struct {
	Person
	specialty string
	hospital  *Hospital   // se asigna al contratarlo con Hospital.Hire
	available atomic.Bool // atómico: se puede cambiar mientras el hospital despacha
}

// NewDoctor valida los datos y crea un doctor disponible.
func NewDoctor(id, name string, age int, specialty string) (*Doctor, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("%w: id and name are required", ErrInvalidStaff)
	}
	if age < 0 {
		return nil, fmt.Errorf("%w: age %d is negative", ErrInvalidStaff, age)
	}
	if strings.TrimSpace(specialty) == "" {
		return nil, fmt.Errorf("%w: specialty is required", ErrInvalidStaff)
	}
	d := &Doctor{Person: newPerson(id, name, age), specialty: specialty}
	d.available.Store(true)
	return d, nil
}

// Specialty devuelve la especialidad del doctor.
func (d *Doctor) Specialty() string { return d.specialty }

// DiagnosePatient deja al doctor como médico asignado del paciente.
// Solo existe en Doctor: el camillero no puede diagnosticar.
func (d *Doctor) DiagnosePatient(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	p.assignDoctor(d)
	return nil
}

// AttendSleepEmergency atiende una emergencia de sueño en el lugar donde está
// el paciente. Reutiliza Attend y descarta el registro: devuelve solo el error.
func (d *Doctor) AttendSleepEmergency(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	_, err := d.Attend(p, p.Location())
	return err
}

// Available indica si el doctor puede atender un nuevo episodio.
func (d *Doctor) Available() bool { return d.available.Load() }

// SetAvailable marca al doctor como disponible o no (por ejemplo, en descanso).
func (d *Doctor) SetAvailable(v bool) { d.available.Store(v) }

// setHospital lo llama Hospital.Hire al contratar al doctor.
func (d *Doctor) setHospital(h *Hospital) { d.hospital = h }

// Attend atiende un ataque de sueño del paciente en location: lo duerme en el
// pasillo, intenta darle sala y registra el episodio en el hospital.
func (d *Doctor) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if d.hospital == nil {
		return EpisodeRecord{}, fmt.Errorf("%w: doctor %s", ErrNotHired, d.ID())
	}
	return d.hospital.handleSleepAttack(d, p, location)
}

// MyEpisodes devuelve solo los episodios del historial que atendió este doctor.
func (d *Doctor) MyEpisodes() []EpisodeRecord {
	var mine []EpisodeRecord
	if d.hospital == nil {
		return mine
	}
	d.hospital.mu.Lock()
	defer d.hospital.mu.Unlock()
	for _, ep := range d.hospital.history {
		if ep.attendedBy == Attender(d) {
			mine = append(mine, ep)
		}
	}
	return mine
=======
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
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
}
