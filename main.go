<<<<<<< HEAD
// Programa principal: arma el escenario de la sección 6 e imprime los
// resultados. La lógica de negocio vive en el paquete hospital.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
=======
package main

import (
	"fmt"
	"log"
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d

	"costenos-narcolepsia/hospital"
)

<<<<<<< HEAD
// check termina el programa si el armado del escenario falla.
func check(err error) {
	if err != nil {
		fmt.Println("error armando el escenario:", err)
		os.Exit(1)
	}
}

// printSevereReport imprime SevereReport. El mapa no tiene orden: se ordenan
// los pacientes por ID para que la salida sea siempre la misma.
func printSevereReport(h *hospital.Hospital) {
	report := h.SevereReport()
	severe := make([]*hospital.Patient, 0, len(report))
	for p := range report {
		severe = append(severe, p)
	}
	sort.Slice(severe, func(i, j int) bool { return severe[i].ID() < severe[j].ID() })
	for _, p := range severe {
		fmt.Printf("%s %s: %d episodio(s)\n", p.ID(), p.Name(), report[p])
=======
func check(err error) {
	if err != nil {
		log.Fatal(err)
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
	}
}

func main() {
<<<<<<< HEAD
	sim := flag.Bool("sim", false, "ejecuta la simulación concurrente en vez del escenario")
	flag.Parse()
	if *sim {
		runSimulation()
		return
	}

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
=======
	// 1. Crear el hospital.
	h := hospital.NewHospital("Hospital Costeño")

	fmt.Println("===================================")
	fmt.Println(" DEMOSTRACIÓN DEL HOSPITAL")
	fmt.Println("===================================")
	fmt.Println("Hospital:", h.Name())

	// 2. Crear y contratar dos doctores.
	doctor1 := hospital.NewDoctor(
		hospital.NewPerson("DOC-001", "Ana", 40),
		h,
	)

	doctor2 := hospital.NewDoctor(
		hospital.NewPerson("DOC-002", "Carlos", 45),
		h,
	)

	check(h.HireDoctor(doctor1))
	check(h.HireDoctor(doctor2))

	// 3. Crear y contratar un camillero.
	orderly := hospital.NewOrderly(
		"CAM-001",
		"Pedro",
		30,
		h,
	)

	check(h.HireOrderly(orderly))

	fmt.Println("\nPersonal registrado:")
	fmt.Println("-", doctor1.Name())
	fmt.Println("-", doctor2.Name())
	fmt.Println("-", orderly.Name())

	// 4. Registrar tres habitaciones con capacidad para un paciente.
	for i := 1; i <= 3; i++ {
		room, err := hospital.NewRoom(i, 1)
		check(err)
		check(h.AddRoom(room))
	}

	fmt.Println("\nHabitaciones registradas:", len(h.Rooms()))

	// 5. Crear y admitir cinco pacientes.
	patients := make([]*hospital.Patient, 0, 5)

	patientData := []struct {
		id    string
		name  string
		age   int
		level hospital.NarcolepsyLevel
	}{
		{"PAC-001", "Luis", 20, hospital.Mild},
		{"PAC-002", "María", 25, hospital.Moderate},
		{"PAC-003", "José", 32, hospital.Severe},
		{"PAC-004", "Andrés", 28, hospital.Severe},
		{"PAC-005", "Sofía", 22, hospital.Mild},
	}

	for _, data := range patientData {
		p, err := hospital.NewPatient(
			data.id,
			data.name,
			data.age,
			data.level,
		)
		check(err)

>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
		check(h.AdmitPatient(p))
		patients = append(patients, p)
	}

<<<<<<< HEAD
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

	fmt.Println("\n== 5.4 Reporte de severidad ==")
	printSevereReport(h)

	fmt.Println("\n== Información adicional: salas ==")
	for _, r := range h.Rooms() {
		fmt.Println(r)
=======
	fmt.Println("\nPacientes admitidos:", len(h.Patients()))

	// 6. Diagnosticar a los pacientes con los doctores.
	for i, p := range patients {
		if i%2 == 0 {
			check(doctor1.DiagnosePatient(p))
		} else {
			check(doctor2.DiagnosePatient(p))
		}
	}

	// 7. Simular cuatro ataques de sueño.
	// Los primeros tres pacientes ocupan las tres habitaciones.
	// El cuarto se queda en el pasillo porque no hay habitaciones libres.
	attenders := []hospital.Attender{
		doctor1,
		doctor2,
		orderly,
		doctor1,
	}

	locations := []string{
		"cafetería",
		"biblioteca",
		"recepción",
		"pasillo principal",
	}

	fmt.Println("\n--- Simulación de ataques ---")

	for i := 0; i < 4; i++ {
		p := patients[i]
		attender := attenders[i]

		episode, err := attender.Attend(p, locations[i])

		fmt.Println("\nPaciente:", p.Name())
		fmt.Println("Personal:", attender.Name())
		fmt.Println("Episodio:", episode.Summary())

		if err != nil {
			fmt.Println("Aviso durante la atención:", err)
		}
	}

	// 8. Consultar quién sigue dormido en el pasillo.
	fmt.Println("\n--- Consulta 1: pacientes en el pasillo ---")

	hallway := h.PatientsInHallway()

	if len(hallway) == 0 {
		fmt.Println("No hay pacientes en el pasillo.")
	}

	for _, p := range hallway {
		fmt.Println("-", p)
	}

	// 9. Despertar al primer paciente, liberando su habitación.
	fmt.Println("\n--- Despertar a un paciente ---")

	check(patients[0].WakeUp())
	fmt.Println(patients[0].Name(), "se ha despertado y liberó su habitación.")

	// 10. El camillero intenta asignar una habitación al paciente
	// que había quedado dormido en el pasillo.
	fmt.Println("\n--- Reasignación de habitación ---")

	episode, err := orderly.Attend(patients[3], "pasillo principal")
	fmt.Println("Episodio:", episode.Summary())

	if err != nil {
		fmt.Println("Aviso durante la reasignación:", err)
	} else {
		fmt.Println("Habitación asignada correctamente.")
	}

	// 11. Consulta 2: estado de las habitaciones.
	fmt.Println("\n--- Consulta 2: habitaciones ---")

	for _, room := range h.Rooms() {
		fmt.Println("-", room)

		for _, p := range room.Occupants() {
			fmt.Println("  Ocupante:", p.Name())
		}
	}

	// 12. Consulta 3: informe de pacientes con nivel severo.
	fmt.Println("\n--- Consulta 3: informe de casos severos ---")

	report := h.SevereReport()

	if len(report) == 0 {
		fmt.Println("No hay pacientes con nivel severo.")
	}

	for p, count := range report {
		fmt.Printf("%s: %d episodio(s) registrado(s)\n",
			p.Name(), count)
	}

	// 13. Consulta 4: episodios registrados por cada miembro del personal.
	fmt.Println("\n--- Consulta 4: episodios por personal ---")

	printEpisodes(doctor1.Name(), doctor1.MyEpisodes())
	printEpisodes(doctor2.Name(), doctor2.MyEpisodes())
	printEpisodes(orderly.Name(), orderly.MyEpisodes())

	fmt.Println("\n--- Fin de la demostración ---")
}

func printEpisodes(name string, episodes []hospital.EpisodeRecord) {
	fmt.Println("\nPersonal:", name)

	if len(episodes) == 0 {
		fmt.Println("Sin episodios registrados.")
		return
	}

	for _, episode := range episodes {
		fmt.Println("-", episode.Summary())
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
	}
}
