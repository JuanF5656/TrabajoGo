package hospital

import (
	"fmt"
	"time"
)

// EpisodeRecord almacena los datos de un episodio de narcolepsia.
type EpisodeRecord struct {
	id         string
	at         time.Time
	patient    *Patient
	attendedBy Attender
	location   string
	room       *Room // nil si no se consiguió habitación
}

// NewEpisodeRecord crea un registro de episodio.
func NewEpisodeRecord(
	id string,
	at time.Time,
	patient *Patient,
	attendedBy Attender,
	location string,
	room *Room,
) EpisodeRecord {
	return EpisodeRecord{
		id:         id,
		at:         at,
		patient:    patient,
		attendedBy: attendedBy,
		location:   location,
		room:       room,
	}
}

// ID devuelve el identificador del episodio.
func (e EpisodeRecord) ID() string {
	return e.id
}

// At devuelve la fecha y hora del episodio.
func (e EpisodeRecord) At() time.Time {
	return e.at
}

// Patient devuelve el paciente del episodio.
func (e EpisodeRecord) Patient() *Patient {
	return e.patient
}

// AttendedBy devuelve el miembro del personal que atendió el episodio.
func (e EpisodeRecord) AttendedBy() Attender {
	return e.attendedBy
}

// Location devuelve el lugar donde ocurrió el episodio.
func (e EpisodeRecord) Location() string {
	return e.location
}

// Room devuelve la habitación asignada, o nil si no hubo una.
func (e EpisodeRecord) Room() *Room {
	return e.room
}

// Summary devuelve un resumen legible del episodio.
func (e EpisodeRecord) Summary() string {
	patientID := "desconocido"
	if e.patient != nil {
		patientID = e.patient.ID()
	}

	attenderName := "desconocido"
	if e.attendedBy != nil {
		attenderName = e.attendedBy.Name()
	}

	roomDescription := "Sin habitación"
	if e.room != nil {
		roomDescription = fmt.Sprintf(
			"Habitación %d",
			e.room.Number(),
		)
	}

	return fmt.Sprintf(
		"%s | %s | Paciente: %s | Personal: %s | Lugar: %s | %s",
		e.id,
		e.at.Format("02/01/2006 15:04:05"),
		patientID,
		attenderName,
		e.location,
		roomDescription,
	)
}
