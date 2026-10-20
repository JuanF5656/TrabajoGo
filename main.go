// Programa principal: arma el escenario de la sección 6 e imprime los
// resultados. La lógica de negocio vive en el paquete hospital.
package main

import (
	"fmt"
	"os"
	"sort"

	"costenos-narcolepsia/hospital"
)

// check termina el programa si el armado del escenario falla.
func check(err error) {
	if err != nil {
		fmt.Println("error armando el escenario:", err)
		os.Exit(1)
	}
}

func main() {
	h := hospital.NewHospital("Hospital de los Costeños con Narcolepsia")

	// Personal: 2 doctores y 1 camillero, todos guardados como Attender.
	d1, err := hospital.NewDoctor("D-001", "Dra. Salcedo", 45, "Neurologia")
	check(err)
	d2, err := hospital.NewDoctor("D-002", "Dr. Mendoza", 38, "Medicina del sueño")
	check(err)
	o1, err := hospital.NewOrderly("O-001", "Camillero Pacheco", 30)
	check(err)
	for _, a := range []hospital.Attender{d1, o1, d2} {
		check(h.Hire(a))
	}

	// 3 salas de capacidad 1.
	for n := 101; n <= 103; n++ {
		r, err := hospital.NewRoom(n, 1)
		check(err)
		check(h.AddRoom(r))
	}

	// 5 pacientes (3 Severe). El último no tiene ataque.
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
		p, err := hospital.NewPatient(x.id, x.name, 40, x.level)
		check(err)
		check(h.AdmitPatient(p))
		patients = append(patients, p)
	}

	// 4 ataques en lugares distintos: 3 consiguen sala, el 4.º no.
	fmt.Println("== Ataques de sueño ==")
	var noRoomErr error
	for i, x := range people[:4] {
		ep, err := h.DispatchSleepAttack(patients[i], x.place)
		fmt.Println(ep.Summary())
		if err != nil {
			fmt.Println("  error:", err)
			noRoomErr = err
		}
	}
	waiting := patients[3]

	fmt.Println("\n== 5.1 Pacientes dormidos en el pasillo ==")
	for _, p := range h.PatientsInHallway() {
		fmt.Printf("%s en %s\n", p, p.Location())
	}

	for _, d := range []*hospital.Doctor{d1, d2} {
		fmt.Printf("\n== 5.2 Episodios atendidos por %s ==\n", d.Name())
		for _, ep := range d.MyEpisodes() {
			fmt.Println(ep.Summary())
		}
	}

	fmt.Printf("\n== 5.3 Asignación de sala a %s ==\n", waiting.ID())
	fmt.Println("Antes de la reasignación:", noRoomErr)
	check(patients[0].WakeUp())
	fmt.Printf("%s se despertó y liberó su sala.\n", patients[0].Name())
	room, err := h.AssignRoom(waiting)
	if err != nil {
		fmt.Println("Después de la reasignación:", err)
	} else {
		fmt.Printf("Después de la reasignación: %s ahora está en %s\n", waiting.Name(), room)
	}

	// El mapa no tiene orden: se ordenan los pacientes por ID para que la
	// salida sea siempre la misma.
	fmt.Println("\n== 5.4 Reporte de severidad ==")
	report := h.SevereReport()
	severe := make([]*hospital.Patient, 0, len(report))
	for p := range report {
		severe = append(severe, p)
	}
	sort.Slice(severe, func(i, j int) bool { return severe[i].ID() < severe[j].ID() })
	for _, p := range severe {
		fmt.Printf("%s %s: %d episodio(s)\n", p.ID(), p.Name(), report[p])
	}

	fmt.Println("\n== Información adicional: salas ==")
	for _, r := range h.Rooms() {
		fmt.Println(r)
	}
}
