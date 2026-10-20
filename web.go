package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"costenos-narcolepsia/hospital"
)

type patientView struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Age        int    `json:"age"`
	Level      string `json:"level"`
	State      string `json:"state"`
	Location   string `json:"location"`
	RoomNumber int    `json:"roomNumber"`
}

type roomView struct {
	Number       int      `json:"number"`
	State        string   `json:"state"`
	Capacity     int      `json:"capacity"`
	Occupied     int      `json:"occupied"`
	PatientNames []string `json:"patientNames"`
}

type severeView struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Episodes int    `json:"episodes"`
}

type dashboardView struct {
	Patients []patientView `json:"patients"`
	Rooms    []roomView    `json:"rooms"`
	Severe   []severeView  `json:"severe"`
}

func startWeb(h *hospital.Hospital) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		data := dashboardView{
			Patients: make([]patientView, 0),
			Rooms:    make([]roomView, 0),
			Severe:   make([]severeView, 0),
		}

		severeReport := h.SevereReport()

		for _, p := range h.Patients() {
			patient := patientView{
				ID:       p.ID(),
				Name:     p.Name(),
				Age:      p.Age(),
				Level:    fmt.Sprint(p.Level()),
				State:    fmt.Sprint(p.State()),
				Location: p.Location(),
			}

			if room := p.Room(); room != nil {
				patient.RoomNumber = room.Number()
			}

			data.Patients = append(data.Patients, patient)

			if episodes := severeReport[p]; episodes > 0 {
				data.Severe = append(data.Severe, severeView{
					ID:       p.ID(),
					Name:     p.Name(),
					Episodes: episodes,
				})
			}
		}

		for _, room := range h.Rooms() {
			names := make([]string, 0)

			for _, p := range room.Occupants() {
				names = append(names, p.Name())
			}

			data.Rooms = append(data.Rooms, roomView{
				Number:       room.Number(),
				State:        fmt.Sprint(room.State()),
				Capacity:     room.Capacity(),
				Occupied:     len(room.Occupants()),
				PatientNames: names,
			})
		}

		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Si el cliente ya se desconectó, no hay nada más que enviar.
			return
		}
	})

	mux.Handle("/", http.FileServer(http.Dir("./web")))

	fmt.Println("Interfaz disponible en http://localhost:8080")
	return http.ListenAndServe(":8080", mux)
}
