package main

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
)

// newTestServer crea un servidor con el escenario inicial y un sistema de
// archivos falso para la página (los tests solo usan /api).
func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	s, err := newServer()
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	var static fs.FS = fstest.MapFS{"index.html": {Data: []byte("<html></html>")}}
	return s.routes(static)
}

// do hace una petición y devuelve el código HTTP y el cuerpo ya decodificado.
func do(t *testing.T, h http.Handler, method, path, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("%s %s: respuesta no es JSON (%q): %v", method, path, rec.Body.String(), err)
	}
	return rec.Code, out
}

func attack(t *testing.T, h http.Handler, id string) {
	t.Helper()
	code, out := do(t, h, http.MethodPost, "/api/attack", `{"patientId":"`+id+`","location":"lugar de `+id+`"}`)
	if code != http.StatusOK {
		t.Fatalf("attack %s: %d %v", id, code, out)
	}
}

func TestState_FivePatients(t *testing.T) {
	h := newTestServer(t)
	code, out := do(t, h, http.MethodGet, "/api/state", "")
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	if n := len(out["patients"].([]any)); n != 5 {
		t.Errorf("patients = %d, want 5", n)
	}
	if n := len(out["rooms"].([]any)); n != 3 {
		t.Errorf("rooms = %d, want 3", n)
	}
	types := map[string]int{}
	for _, s := range out["staff"].([]any) {
		types[s.(map[string]any)["type"].(string)]++
	}
	if types["Doctor"] != 2 || types["Orderly"] != 1 {
		t.Errorf("staff types = %v, want 2 Doctor and 1 Orderly", types)
	}
}

func TestFourAttacks_OneInHallway_AssignConflict(t *testing.T) {
	h := newTestServer(t)
	for _, id := range []string{"P-001", "P-002", "P-003", "P-004"} {
		attack(t, h, id)
	}

	_, q := do(t, h, http.MethodGet, "/api/queries", "")
	hallway := q["hallway"].([]any)
	if len(hallway) != 1 || hallway[0].(map[string]any)["id"] != "P-004" {
		t.Fatalf("hallway = %v, want only P-004", hallway)
	}

	code, out := do(t, h, http.MethodPost, "/api/assign", `{"patientId":"P-004"}`)
	if code != http.StatusConflict {
		t.Fatalf("assign status = %d, want 409", code)
	}
	if msg, _ := out["error"].(string); !strings.Contains(msg, "no room available") {
		t.Errorf("error = %q", msg)
	}
}

func TestWakeUpThenAssign_OK(t *testing.T) {
	h := newTestServer(t)
	for _, id := range []string{"P-001", "P-002", "P-003", "P-004"} {
		attack(t, h, id)
	}
	if code, out := do(t, h, http.MethodPost, "/api/wakeup", `{"patientId":"P-001"}`); code != http.StatusOK {
		t.Fatalf("wakeup: %d %v", code, out)
	}
	code, out := do(t, h, http.MethodPost, "/api/assign", `{"patientId":"P-004"}`)
	if code != http.StatusOK {
		t.Fatalf("assign: %d %v", code, out)
	}
	if out["room"] != float64(101) {
		t.Errorf("room = %v, want 101", out["room"])
	}

	// El paciente reasignado cuenta 1 episodio: AssignRoom no crea episodios.
	_, q := do(t, h, http.MethodGet, "/api/queries", "")
	for _, s := range q["severeReport"].([]any) {
		m := s.(map[string]any)
		if m["id"] == "P-004" && m["episodes"] != float64(1) {
			t.Errorf("P-004 episodes = %v, want 1", m["episodes"])
		}
	}
}

func TestErrorCodes(t *testing.T) {
	h := newTestServer(t)
	attack(t, h, "P-001")

	cases := []struct {
		name, method, path, body string
		want                     int
	}{
		{"bad json", http.MethodPost, "/api/attack", `{`, http.StatusBadRequest},
		{"missing id", http.MethodPost, "/api/wakeup", `{}`, http.StatusBadRequest},
		{"unknown patient", http.MethodPost, "/api/assign", `{"patientId":"P-999"}`, http.StatusNotFound},
		{"already asleep", http.MethodPost, "/api/attack", `{"patientId":"P-001"}`, http.StatusConflict},
		{"already awake", http.MethodPost, "/api/wakeup", `{"patientId":"P-005"}`, http.StatusConflict},
		{"not in hallway", http.MethodPost, "/api/assign", `{"patientId":"P-005"}`, http.StatusConflict},
		{"wrong method", http.MethodGet, "/api/attack", "", http.StatusMethodNotAllowed},
	}
	for _, c := range cases {
		code, out := do(t, h, c.method, c.path, c.body)
		if code != c.want {
			t.Errorf("%s: status = %d, want %d (%v)", c.name, code, c.want, out)
		}
		if _, ok := out["error"]; !ok {
			t.Errorf("%s: response has no error field", c.name)
		}
	}
}

func TestReset_RestoresInitialScenario(t *testing.T) {
	h := newTestServer(t)
	for _, id := range []string{"P-001", "P-002"} {
		attack(t, h, id)
	}
	if code, _ := do(t, h, http.MethodPost, "/api/reset", ""); code != http.StatusOK {
		t.Fatalf("reset status = %d", code)
	}
	_, out := do(t, h, http.MethodGet, "/api/state", "")
	for _, p := range out["patients"].([]any) {
		if st := p.(map[string]any)["state"]; st != "Awake" {
			t.Errorf("after reset a patient is %v, want Awake", st)
		}
	}
	_, q := do(t, h, http.MethodGet, "/api/queries", "")
	if q["lastAssignment"] != nil {
		t.Errorf("lastAssignment should be cleared, got %v", q["lastAssignment"])
	}
}

// Peticiones simultáneas: bajo -race comprueba que el mutex protege el hospital.
func TestConcurrentRequests(t *testing.T) {
	h := newTestServer(t)
	ids := []string{"P-001", "P-002", "P-003", "P-004", "P-005"}
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				do(t, h, http.MethodPost, "/api/attack", `{"patientId":"`+id+`","location":"x"}`)
				do(t, h, http.MethodGet, "/api/state", "")
				do(t, h, http.MethodGet, "/api/queries", "")
				do(t, h, http.MethodPost, "/api/assign", `{"patientId":"`+id+`"}`)
				do(t, h, http.MethodPost, "/api/wakeup", `{"patientId":"`+id+`"}`)
			}
		}(id)
	}
	wg.Wait()
}
