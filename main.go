package main

import (
	"fmt"
	"log"

	"costenos-narcolepsia/hospital"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
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

		check(h.AdmitPatient(p))
		patients = append(patients, p)
	}

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
	}
}
