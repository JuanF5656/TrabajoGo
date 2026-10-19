package hospital

import "time"

// =====================================================================
// STUB DE CONTRATO — Persona B completa este archivo.
// A depende solo de: el campo patient y el getter Patient()
// (los usa SevereReport).
// =====================================================================

// EpisodeRecord — TODO(B): constructor, getters restantes y Summary().
type EpisodeRecord struct {
	id         string
	at         time.Time
	patient    *Patient
	attendedBy Attender
	location   string
	room       *Room // nil si no se consiguió sala
}

// Patient devuelve el paciente que tuvo el episodio.
func (e EpisodeRecord) Patient() *Patient { return e.patient }
