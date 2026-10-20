package hospital

import "fmt"

type Hospital struct {
	name     string
	staff    []Attender // Persona B: doctores y camilleros, despachados por interfaz
	patients []*Patient
	rooms    []*Room
	history  []EpisodeRecord // todos los episodios registrados (Persona B los agrega)
}

func NewHospital(name string) *Hospital {
	return &Hospital{name: name}
}

func (h *Hospital) Name() string { return h.name }

// Patients y Rooms devuelven copias para que main pueda recorrerlos
// sin poder alterar las listas internas del hospital.
func (h *Hospital) Patients() []*Patient {
	out := make([]*Patient, len(h.patients))
	copy(out, h.patients)
	return out
}

func (h *Hospital) Rooms() []*Room {
	out := make([]*Room, len(h.rooms))
	copy(out, h.rooms)
	return out
}

// AddRoom registra una sala. No se permiten números repetidos.
func (h *Hospital) AddRoom(r *Room) error {
	if r == nil {
		return fmt.Errorf("%w: room is nil", ErrInvalidRoom)
	}
	for _, existing := range h.rooms {
		if existing.number == r.number {
			return fmt.Errorf("%w: %d", ErrRoomAlreadyRegistered, r.number)
		}
	}
	h.rooms = append(h.rooms, r)
	return nil
}

// AdmitPatient ingresa un paciente. No se admite dos veces el mismo id.
func (h *Hospital) AdmitPatient(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if h.findPatient(p.ID()) != nil {
		return fmt.Errorf("%w: %s", ErrPatientAlreadyAdmitted, p.ID())
	}
	h.patients = append(h.patients, p)
	return nil
}

func (h *Hospital) findPatient(id string) *Patient {
	for _, p := range h.patients {
		if p.ID() == id {
			return p
		}
	}
	return nil
}

func (h *Hospital) isAdmitted(p *Patient) bool {
	for _, q := range h.patients {
		if q == p {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------
// Consultas de la sección 5 (devuelven datos; imprimir es trabajo de main)
// ---------------------------------------------------------------------

// AssignRoom (5.3) busca la PRIMERA sala disponible, la ocupa y pasa al
// paciente a AsleepInBed. Si no hay sala, devuelve ErrNoRoomAvailable y el
// paciente sigue en AsleepInHallway: no hay panic ni sala inventada.
func (h *Hospital) AssignRoom(p *Patient) (*Room, error) {
	if p == nil {
		return nil, ErrNilPatient
	}
	if !h.isAdmitted(p) {
		return nil, fmt.Errorf("%w: %s", ErrPatientNotAdmitted, p.ID())
	}
	if p.State() != AsleepInHallway {
		return nil, fmt.Errorf("%w: %s is %s", ErrNotInHallway, p.ID(), p.State())
	}
	for _, r := range h.rooms {
		if !r.IsAvailable() {
			continue
		}
		if err := r.Occupy(p); err != nil {
			return nil, fmt.Errorf("assigning room %d to %s: %w", r.Number(), p.ID(), err)
		}
		p.moveToBed(r)
		return r, nil
	}
	return nil, fmt.Errorf("%w: patient %s stays in the hallway", ErrNoRoomAvailable, p.ID())
}

// PatientsInHallway (5.1) devuelve los pacientes dormidos en un pasillo,
// en el orden en que fueron admitidos, para despachar a un camillero.
func (h *Hospital) PatientsInHallway() []*Patient {
	var result []*Patient
	for _, p := range h.patients {
		if p.State() == AsleepInHallway {
			result = append(result, p)
		}
	}
	return result
}

func (h *Hospital) SevereReport() map[*Patient]int {
	report := make(map[*Patient]int)
	for _, p := range h.patients {
		if p.Level() == Severe {
			report[p] = 0
		}
	}
	for _, ep := range h.history {
		if _, isSevere := report[ep.Patient()]; isSevere {
			report[ep.Patient()]++
		}
	}
	return report
}

// ---------------------------------------------------------------------
// Historial de episodios
// ---------------------------------------------------------------------

// nextEpisodeID genera el siguiente id ("EP-001", "EP-002", ...). Vive aquí
// y no en Doctor ni en Orderly para que la numeración tenga un solo dueño:
// el hospital, que es quien guarda el historial.
func (h *Hospital) nextEpisodeID() string {
	return fmt.Sprintf("EP-%03d", len(h.history)+1)
}

// RegisterEpisode agrega un episodio al historial. Solo acepta episodios de
// pacientes admitidos en este hospital, para que SevereReport no cuente
// episodios de gente que no es de aquí.
func (h *Hospital) RegisterEpisode(ep EpisodeRecord) error {
	if ep.patient == nil {
		return ErrNilPatient
	}
	if !h.isAdmitted(ep.patient) {
		return fmt.Errorf("%w: %s", ErrPatientNotAdmitted, ep.patient.ID())
	}
	h.history = append(h.history, ep)
	return nil
}
