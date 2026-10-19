package hospital

import (
	"errors"
	"testing"
)

// --- helpers ---

func mustPatient(t *testing.T, id, name string, level NarcolepsyLevel) *Patient {
	t.Helper()
	p, err := NewPatient(id, name, 40, level)
	if err != nil {
		t.Fatalf("NewPatient(%s): %v", id, err)
	}
	return p
}

// newTestHospital crea un hospital con nRooms salas de capacidad 1
// (números 101, 102, ...) y admite los pacientes dados.
func newTestHospital(t *testing.T, nRooms int, patients ...*Patient) *Hospital {
	t.Helper()
	h := NewHospital("Hospital de prueba")
	for i := 0; i < nRooms; i++ {
		r, err := NewRoom(101+i, 1)
		if err != nil {
			t.Fatalf("NewRoom: %v", err)
		}
		if err := h.AddRoom(r); err != nil {
			t.Fatalf("AddRoom: %v", err)
		}
	}
	for _, p := range patients {
		if err := h.AdmitPatient(p); err != nil {
			t.Fatalf("AdmitPatient(%s): %v", p.ID(), err)
		}
	}
	return h
}

// --- 5.3 AssignRoom: hay sala libre (test obligatorio 1) ---

func TestAssignRoom_FreeRoom(t *testing.T) {
	p := mustPatient(t, "P-001", "Yeimy Padilla", Severe)
	h := newTestHospital(t, 2, p)

	if err := p.SufferSleepAttack("cafeteria"); err != nil {
		t.Fatalf("SufferSleepAttack: %v", err)
	}
	room, err := h.AssignRoom(p)
	if err != nil {
		t.Fatalf("AssignRoom returned error: %v", err)
	}
	if room.Number() != 101 {
		t.Errorf("expected first available room 101, got %d", room.Number())
	}
	if room.State() != Occupied {
		t.Errorf("room state = %s, want Occupied", room.State())
	}
	if p.State() != AsleepInBed {
		t.Errorf("patient state = %s, want AsleepInBed", p.State())
	}
	if p.Room() != room {
		t.Errorf("patient room not set")
	}
}

// --- 5.3 AssignRoom: no hay sala (test obligatorio 2) ---

func TestAssignRoom_NoRoomAvailable(t *testing.T) {
	p1 := mustPatient(t, "P-001", "Yeimy Padilla", Severe)
	p2 := mustPatient(t, "P-002", "Wilfrido Berrio", Severe)
	h := newTestHospital(t, 1, p1, p2)

	_ = p1.SufferSleepAttack("cafeteria")
	if _, err := h.AssignRoom(p1); err != nil {
		t.Fatalf("first AssignRoom should succeed: %v", err)
	}

	_ = p2.SufferSleepAttack("hallway 2")
	room, err := h.AssignRoom(p2)
	if !errors.Is(err, ErrNoRoomAvailable) {
		t.Fatalf("expected ErrNoRoomAvailable, got %v", err)
	}
	if room != nil {
		t.Errorf("expected nil room, got %v", room)
	}
	if p2.State() != AsleepInHallway {
		t.Errorf("patient should stay AsleepInHallway, got %s", p2.State())
	}
	if p2.Location() != "hallway 2" {
		t.Errorf("location changed to %q", p2.Location())
	}
}

// Escenario de la sección 6: alguien despierta, libera la sala y el
// paciente del pasillo la recibe.
func TestAssignRoom_AfterWakeUpReleasesRoom(t *testing.T) {
	p1 := mustPatient(t, "P-001", "Yeimy Padilla", Severe)
	p2 := mustPatient(t, "P-002", "Wilfrido Berrio", Severe)
	h := newTestHospital(t, 1, p1, p2)

	_ = p1.SufferSleepAttack("cafeteria")
	first, _ := h.AssignRoom(p1)
	_ = p2.SufferSleepAttack("hallway 2")
	if _, err := h.AssignRoom(p2); err == nil {
		t.Fatal("expected error before wake up")
	}

	if err := p1.WakeUp(); err != nil {
		t.Fatalf("WakeUp: %v", err)
	}
	if !first.IsAvailable() || first.State() != Available {
		t.Fatalf("room should be available after wake up, got %s", first)
	}

	room, err := h.AssignRoom(p2)
	if err != nil {
		t.Fatalf("AssignRoom after release: %v", err)
	}
	if room != first {
		t.Errorf("expected the released room %d, got %d", first.Number(), room.Number())
	}
}

func TestAssignRoom_InvalidCases(t *testing.T) {
	p := mustPatient(t, "P-001", "Yeimy Padilla", Mild)
	outsider := mustPatient(t, "P-999", "No admitido", Mild)
	h := newTestHospital(t, 1, p)

	if _, err := h.AssignRoom(nil); !errors.Is(err, ErrNilPatient) {
		t.Errorf("nil patient: got %v", err)
	}
	if _, err := h.AssignRoom(p); !errors.Is(err, ErrNotInHallway) {
		t.Errorf("awake patient: got %v", err)
	}
	_ = outsider.SufferSleepAttack("parking")
	if _, err := h.AssignRoom(outsider); !errors.Is(err, ErrPatientNotAdmitted) {
		t.Errorf("not admitted: got %v", err)
	}
}

// --- 5.1 PatientsInHallway (test obligatorio 3) ---

func TestPatientsInHallway(t *testing.T) {
	a := mustPatient(t, "P-001", "A", Mild)
	b := mustPatient(t, "P-002", "B", Moderate)
	c := mustPatient(t, "P-003", "C", Severe)
	h := newTestHospital(t, 1, a, b, c)

	if got := h.PatientsInHallway(); len(got) != 0 {
		t.Fatalf("expected nobody in hallway, got %d", len(got))
	}

	_ = a.SufferSleepAttack("cafeteria")
	_, _ = h.AssignRoom(a) // a pasa a cama
	_ = c.SufferSleepAttack("radiology queue")

	got := h.PatientsInHallway()
	if len(got) != 1 || got[0] != c {
		t.Fatalf("expected only P-003 in hallway, got %v", got)
	}
}

// --- 5.4 SevereReport ---

func TestSevereReport(t *testing.T) {
	mild := mustPatient(t, "P-001", "Mild", Mild)
	s1 := mustPatient(t, "P-002", "Severe uno", Severe)
	s2 := mustPatient(t, "P-003", "Severe dos", Severe)
	h := newTestHospital(t, 0, mild, s1, s2)

	// RegisterEpisode es de Persona B; mientras tanto el test (que está
	// dentro del paquete) carga el historial directamente.
	h.history = []EpisodeRecord{
		{patient: s1}, {patient: s1}, {patient: s1},
		{patient: mild},
	}

	report := h.SevereReport()
	if len(report) != 2 {
		t.Fatalf("expected 2 severe patients, got %d", len(report))
	}
	if report[s1] != 3 {
		t.Errorf("P-002 episodes = %d, want 3", report[s1])
	}
	if n, ok := report[s2]; !ok || n != 0 {
		t.Errorf("P-003 should appear with 0 episodes, got %d (present=%v)", n, ok)
	}
	if _, ok := report[mild]; ok {
		t.Errorf("Mild patient must not appear in the report")
	}
}

// --- Room ---

func TestRoomOccupyAndRelease(t *testing.T) {
	r, _ := NewRoom(101, 1)
	a := mustPatient(t, "P-001", "A", Mild)
	b := mustPatient(t, "P-002", "B", Mild)

	if err := r.Occupy(a); err != nil {
		t.Fatalf("Occupy: %v", err)
	}
	if err := r.Occupy(a); !errors.Is(err, ErrPatientAlreadyInRoom) {
		t.Errorf("double occupy: got %v", err)
	}
	if err := r.Occupy(b); !errors.Is(err, ErrRoomFull) {
		t.Errorf("full room: got %v", err)
	}
	if err := r.Release(b); !errors.Is(err, ErrPatientNotInRoom) {
		t.Errorf("release absent patient: got %v", err)
	}
	if err := r.Release(a); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if r.State() != Available || len(r.Occupants()) != 0 {
		t.Errorf("room should be empty and Available, got %s", r)
	}
}

func TestNewRoomAndPatientValidation(t *testing.T) {
	if _, err := NewRoom(101, 0); !errors.Is(err, ErrInvalidRoom) {
		t.Errorf("capacity 0: got %v", err)
	}
	if _, err := NewPatient("", "Sin id", 30, Mild); !errors.Is(err, ErrInvalidPatient) {
		t.Errorf("empty id: got %v", err)
	}
	if _, err := NewPatient("P-1", "Edad mala", -1, Mild); !errors.Is(err, ErrInvalidPatient) {
		t.Errorf("negative age: got %v", err)
	}
	if _, err := NewPatient("P-1", "Nivel malo", 30, NarcolepsyLevel(9)); !errors.Is(err, ErrInvalidPatient) {
		t.Errorf("invalid level: got %v", err)
	}
}

// --- Patient ---

func TestPatientSleepAndWakeTransitions(t *testing.T) {
	p := mustPatient(t, "P-001", "A", Moderate)

	if err := p.WakeUp(); !errors.Is(err, ErrAlreadyAwake) {
		t.Errorf("wake up awake patient: got %v", err)
	}
	if err := p.SufferSleepAttack("palm tree"); err != nil {
		t.Fatalf("SufferSleepAttack: %v", err)
	}
	if err := p.SufferSleepAttack("again"); !errors.Is(err, ErrAlreadyAsleep) {
		t.Errorf("double attack: got %v", err)
	}
	if err := p.WakeUp(); err != nil {
		t.Fatalf("WakeUp from hallway: %v", err)
	}
	if p.State() != Awake {
		t.Errorf("state = %s, want Awake", p.State())
	}
}

func TestStateStrings(t *testing.T) {
	cases := map[string]string{
		AsleepInHallway.String():  "AsleepInHallway",
		Severe.String():           "Severe",
		Occupied.String():         "Occupied",
		PatientState(42).String(): "UnknownPatientState",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
