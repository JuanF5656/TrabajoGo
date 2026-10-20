package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"sort"
	"sync"

	"costenos-narcolepsia/hospital"
)

// server guarda el hospital y las referencias que la vista necesita.
// Los handlers HTTP corren en goroutines distintas, así que TODO acceso al
// hospital (lectura o escritura) se hace con mu tomado.
type server struct {
	mu         sync.Mutex
	h          *hospital.Hospital
	doctors    []*hospital.Doctor // el paquete no expone el personal: lo guardamos aquí
	staff      []hospital.Attender
	lastAssign *assignDTO
}

// newServer arma el escenario inicial.
func newServer() (*server, error) {
	s := &server{}
	if err := s.resetLocked(); err != nil {
		return nil, err
	}
	return s, nil
}

// resetLocked reconstruye el escenario inicial (sin ataques). Quien lo llame
// debe tener mu tomado, salvo en newServer, donde nadie más ve al servidor aún.
func (s *server) resetLocked() error {
	h := hospital.NewHospital("Hospital de los Costeños con Narcolepsia")

	d1, err := hospital.NewDoctor("D-001", "Dra. Salcedo", 45, "Neurologia")
	if err != nil {
		return err
	}
	d2, err := hospital.NewDoctor("D-002", "Dr. Mendoza", 38, "Medicina del sueño")
	if err != nil {
		return err
	}
	o1, err := hospital.NewOrderly("O-001", "Camillero Pacheco", 30)
	if err != nil {
		return err
	}
	staff := []hospital.Attender{d1, o1, d2}
	for _, a := range staff {
		if err := h.Hire(a); err != nil {
			return err
		}
	}
	for n := 101; n <= 103; n++ {
		r, err := hospital.NewRoom(n, 1)
		if err != nil {
			return err
		}
		if err := h.AddRoom(r); err != nil {
			return err
		}
	}
	people := []struct {
		id, name string
		age      int
		level    hospital.NarcolepsyLevel
	}{
		{"P-001", "Yeimy Padilla", 40, hospital.Severe},
		{"P-002", "Wilfrido Berrio", 52, hospital.Severe},
		{"P-003", "Marelvis Cantillo", 35, hospital.Moderate},
		{"P-004", "Dilson Ospino", 61, hospital.Severe},
		{"P-005", "Ana Beltran", 28, hospital.Mild},
	}
	for _, x := range people {
		p, err := hospital.NewPatient(x.id, x.name, x.age, x.level)
		if err != nil {
			return err
		}
		if err := h.AdmitPatient(p); err != nil {
			return err
		}
	}

	s.h, s.doctors, s.staff, s.lastAssign = h, []*hospital.Doctor{d1, d2}, staff, nil
	return nil
}

// ---------------------------------------------------------------------
// DTOs: lo único que sale como JSON. Nunca se serializa el dominio directo.
// ---------------------------------------------------------------------

type patientDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Level    string `json:"level"`
	State    string `json:"state"`
	Location string `json:"location"`
	Doctor   string `json:"doctor"` // vacío si no tiene doctor asignado
	Room     int    `json:"room"`   // 0 si no tiene sala
}

type roomDTO struct {
	Number    int      `json:"number"`
	Capacity  int      `json:"capacity"`
	State     string   `json:"state"`
	Occupants []string `json:"occupants"` // "P-001 Nombre"
}

type staffDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // "Doctor" u "Orderly"
}

type stateDTO struct {
	Hospital string       `json:"hospital"`
	Patients []patientDTO `json:"patients"`
	Rooms    []roomDTO    `json:"rooms"`
	Staff    []staffDTO   `json:"staff"`
}

type doctorEpisodesDTO struct {
	DoctorID   string   `json:"doctorId"`
	DoctorName string   `json:"doctorName"`
	Episodes   []string `json:"episodes"`
}

type severeDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Episodes int    `json:"episodes"`
}

// assignDTO es el resultado de la última asignación de sala (consulta 5.3).
type assignDTO struct {
	PatientID string `json:"patientId"`
	OK        bool   `json:"ok"`
	Room      int    `json:"room,omitempty"`
	Message   string `json:"message"`
}

type queriesDTO struct {
	Hallway          []patientDTO        `json:"hallway"`
	EpisodesByDoctor []doctorEpisodesDTO `json:"episodesByDoctor"`
	LastAssignment   *assignDTO          `json:"lastAssignment"`
	SevereReport     []severeDTO         `json:"severeReport"`
}

func toPatientDTO(p *hospital.Patient) patientDTO {
	d := patientDTO{
		ID: p.ID(), Name: p.Name(), Age: p.Age(),
		Level: p.Level().String(), State: p.State().String(), Location: p.Location(),
	}
	if doc := p.AssignedDoctor(); doc != nil {
		d.Doctor = doc.Name()
	}
	if r := p.Room(); r != nil {
		d.Room = r.Number()
	}
	return d
}

// staffType descubre el tipo concreto con un type switch. Solo la vista lo
// necesita; el hospital despacha sin saberlo.
func staffType(a hospital.Attender) string {
	switch a.(type) {
	case *hospital.Doctor:
		return "Doctor"
	case *hospital.Orderly:
		return "Orderly"
	default:
		return "Unknown"
	}
}

// ---------------------------------------------------------------------
// Errores HTTP
// ---------------------------------------------------------------------

func statusFor(err error) int {
	switch {
	case errors.Is(err, hospital.ErrNilPatient),
		errors.Is(err, hospital.ErrInvalidPatient),
		errors.Is(err, hospital.ErrInvalidRoom),
		errors.Is(err, hospital.ErrInvalidStaff):
		return http.StatusBadRequest
	case errors.Is(err, hospital.ErrPatientNotAdmitted):
		return http.StatusNotFound
	case errors.Is(err, hospital.ErrAlreadyAsleep),
		errors.Is(err, hospital.ErrAlreadyAwake),
		errors.Is(err, hospital.ErrNotInHallway),
		errors.Is(err, hospital.ErrNoRoomAvailable),
		errors.Is(err, hospital.ErrNoStaffAvailable),
		errors.Is(err, hospital.ErrRoomFull),
		errors.Is(err, hospital.ErrPatientAlreadyInRoom),
		errors.Is(err, hospital.ErrPatientNotInRoom):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// ---------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------

func (s *server) routes(static fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/state", onlyMethod(http.MethodGet, s.handleState))
	mux.HandleFunc("/api/queries", onlyMethod(http.MethodGet, s.handleQueries))
	mux.HandleFunc("/api/attack", onlyMethod(http.MethodPost, s.handleAttack))
	mux.HandleFunc("/api/wakeup", onlyMethod(http.MethodPost, s.handleWakeUp))
	mux.HandleFunc("/api/assign", onlyMethod(http.MethodPost, s.handleAssign))
	mux.HandleFunc("/api/reset", onlyMethod(http.MethodPost, s.handleReset))
	mux.Handle("/", http.FileServer(http.FS(static)))
	return mux
}

func onlyMethod(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h(w, r)
	}
}

func (s *server) handleState(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := stateDTO{Hospital: s.h.Name(), Patients: []patientDTO{}, Rooms: []roomDTO{}, Staff: []staffDTO{}}
	for _, p := range s.h.Patients() {
		st.Patients = append(st.Patients, toPatientDTO(p))
	}
	for _, room := range s.h.Rooms() {
		rd := roomDTO{Number: room.Number(), Capacity: room.Capacity(), State: room.State().String(), Occupants: []string{}}
		for _, o := range room.Occupants() {
			rd.Occupants = append(rd.Occupants, o.ID()+" "+o.Name())
		}
		st.Rooms = append(st.Rooms, rd)
	}
	for _, a := range s.staff {
		st.Staff = append(st.Staff, staffDTO{ID: a.ID(), Name: a.Name(), Type: staffType(a)})
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *server) handleQueries(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	q := queriesDTO{
		Hallway:          []patientDTO{},
		EpisodesByDoctor: []doctorEpisodesDTO{},
		SevereReport:     []severeDTO{},
		LastAssignment:   s.lastAssign,
	}
	for _, p := range s.h.PatientsInHallway() {
		q.Hallway = append(q.Hallway, toPatientDTO(p))
	}
	for _, d := range s.doctors {
		de := doctorEpisodesDTO{DoctorID: d.ID(), DoctorName: d.Name(), Episodes: []string{}}
		for _, ep := range d.MyEpisodes() {
			de.Episodes = append(de.Episodes, ep.Summary())
		}
		q.EpisodesByDoctor = append(q.EpisodesByDoctor, de)
	}
	// El mapa no tiene orden: se ordena por ID para que la respuesta sea estable.
	for p, n := range s.h.SevereReport() {
		q.SevereReport = append(q.SevereReport, severeDTO{ID: p.ID(), Name: p.Name(), Episodes: n})
	}
	sort.Slice(q.SevereReport, func(i, j int) bool { return q.SevereReport[i].ID < q.SevereReport[j].ID })
	writeJSON(w, http.StatusOK, q)
}

type actionRequest struct {
	PatientID string `json:"patientId"`
	Location  string `json:"location"`
}

// patientFor lee el cuerpo JSON y busca al paciente. Si falla ya respondió el
// error y devuelve nil. Requiere mu tomado.
func (s *server) patientFor(w http.ResponseWriter, r *http.Request) (*hospital.Patient, actionRequest) {
	var req actionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return nil, req
	}
	if req.PatientID == "" {
		writeError(w, http.StatusBadRequest, "patientId is required")
		return nil, req
	}
	for _, p := range s.h.Patients() {
		if p.ID() == req.PatientID {
			return p, req
		}
	}
	writeError(w, http.StatusNotFound, fmt.Sprintf("patient %s not found", req.PatientID))
	return nil, req
}

func (s *server) handleAttack(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, req := s.patientFor(w, r)
	if p == nil {
		return
	}
	ep, err := s.h.DispatchSleepAttack(p, req.Location)
	resp := map[string]string{"episode": ep.Summary()}
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, resp)
	case errors.Is(err, hospital.ErrNoRoomAvailable):
		// El episodio sí quedó registrado; solo falta sala. No es un fallo.
		resp["warning"] = err.Error()
		writeJSON(w, http.StatusOK, resp)
	default:
		writeError(w, statusFor(err), err.Error())
	}
}

func (s *server) handleWakeUp(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, _ := s.patientFor(w, r)
	if p == nil {
		return
	}
	if err := s.h.WakeUp(p); err != nil {
		writeError(w, statusFor(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": p.Name() + " se despertó"})
}

func (s *server) handleAssign(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, _ := s.patientFor(w, r)
	if p == nil {
		return
	}
	room, err := s.h.AssignRoom(p)
	if err != nil {
		s.lastAssign = &assignDTO{PatientID: p.ID(), OK: false, Message: err.Error()}
		writeError(w, statusFor(err), err.Error())
		return
	}
	s.lastAssign = &assignDTO{
		PatientID: p.ID(), OK: true, Room: room.Number(),
		Message: fmt.Sprintf("%s ahora está en la sala %d", p.Name(), room.Number()),
	}
	writeJSON(w, http.StatusOK, s.lastAssign)
}

func (s *server) handleReset(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.resetLocked(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "escenario reiniciado"})
}
