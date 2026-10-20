package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"costenos-narcolepsia/hospital"
)

// runSimulation (bonus de concurrencia) lanza cada paciente como una goroutine
// que sufre varios ataques en intervalos aleatorios. Todas usan el mismo
// hospital a la vez; el mutex del hospital evita las condiciones de carrera.
// La salida cambia en cada corrida (por el azar), por eso solo se ejecuta con -sim.
func runSimulation() {
	h := hospital.NewHospital("Hospital de los Costeños con Narcolepsia")

	d1, err := hospital.NewDoctor("D-001", "Dra. Salcedo", 45, "Neurologia")
	check(err)
	d2, err := hospital.NewDoctor("D-002", "Dr. Mendoza", 38, "Medicina del sueño")
	check(err)
	o1, err := hospital.NewOrderly("O-001", "Camillero Pacheco", 30)
	check(err)
	for _, a := range []hospital.Attender{d1, o1, d2} {
		check(h.Hire(a))
	}
	for n := 101; n <= 103; n++ {
		r, err := hospital.NewRoom(n, 1)
		check(err)
		check(h.AddRoom(r))
	}

	const numPatients, attacksEach = 8, 3
	var patients []*hospital.Patient
	for i := 1; i <= numPatients; i++ {
		level := hospital.NarcolepsyLevel(i % 3)
		p, err := hospital.NewPatient(fmt.Sprintf("P-%03d", i), fmt.Sprintf("Paciente %d", i), 30+i, level)
		check(err)
		check(h.AdmitPatient(p))
		patients = append(patients, p)
	}

	fmt.Println("== Simulación concurrente ==")
	var wg sync.WaitGroup
	for i, p := range patients {
		wg.Add(1)
		go func(i int, p *hospital.Patient) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(i)))
			pause := func() { time.Sleep(time.Duration(1+rng.Intn(20)) * time.Millisecond) }
			for a := 1; a <= attacksEach; a++ {
				pause()
				ep, err := h.DispatchSleepAttack(p, fmt.Sprintf("place %d-%d", i+1, a))
				switch {
				case err == nil:
					fmt.Println(ep.Summary())
				case errors.Is(err, hospital.ErrNoRoomAvailable):
					fmt.Println(ep.Summary(), "(esperando sala)")
					// Reintenta unas veces mientras otros pacientes despiertan.
					for try := 0; try < 5; try++ {
						pause()
						if _, err := h.AssignRoom(p); err == nil {
							break
						}
					}
				default:
					fmt.Println(p.ID(), "error:", err)
					return
				}
				pause()
				if err := h.WakeUp(p); err != nil {
					fmt.Println(p.ID(), "error:", err)
					return
				}
			}
		}(i, p)
	}
	wg.Wait()

	fmt.Println("\n== Reporte de severidad (simulación) ==")
	printSevereReport(h)
}
