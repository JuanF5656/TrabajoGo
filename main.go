package main

import (
	"fmt"

	"costenos-narcolepsia/hospital"
)

func main() {
	h := hospital.NewHospital("Hospital de los Costeños con Narcolepsia")

	d1, _ := hospital.NewDoctor("D-001", "Dra. Salcedo", 45)
	d2, _ := hospital.NewDoctor("D-002", "Dr. Mendoza", 38)
	o1, _ := hospital.NewOrderly("O-001", "Camillero Pacheco", 30)
	for _, a := range []hospital.Attender{d1, o1, d2} {
		if err := h.Hire(a); err != nil {
			fmt.Println("error:", err)
		}
	}
	for n := 101; n <= 103; n++ {
		r, _ := hospital.NewRoom(n, 1)
		_ = h.AddRoom(r)
	}
	people := []struct {
		id, name string
		level    hospital.NarcolepsyLevel
		place    string
	}{
		{"P-001", "Yeimy Padilla", hospital.Severe, "cafeteria"},
		{"P-002", "Wilfrido Berrio", hospital.Severe, "parking lot"},
		{"P-003", "Marelvis Cantillo", hospital.Moderate, "radiology queue"},
		{"P-004", "Dilson Ospino", hospital.Severe, "palm tree"},
		{"P-005", "Ana Beltran", hospital.Mild, ""},
	}
	var patients []*hospital.Patient
	for _, x := range people {
		p, _ := hospital.NewPatient(x.id, x.name, 40, x.level)
		_ = h.AdmitPatient(p)
		patients = append(patients, p)
	}

	fmt.Println("== Ataques de sueño ==")
	for i, x := range people[:4] {
		ep, err := h.DispatchSleepAttack(patients[i], x.place)
		fmt.Println(ep.Summary())
		if err != nil {
			fmt.Println("  error:", err)
		}
	}
}
