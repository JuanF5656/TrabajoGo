package hospital

import (
	"fmt"
	"time"
)

<<<<<<< HEAD
// EpisodeRecord es el registro de un ataque de sueño atendido. Es inmutable:
// tiene campos privados y solo getters, así nadie cambia el historial después.
=======
// EpisodeRecord almacena los datos de un episodio de narcolepsia.
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
type EpisodeRecord struct {
	id         string
	at         time.Time
	patient    *Patient
	attendedBy Attender
	location   string
	room       *Room // nil si no se consiguió habitación
}

<<<<<<< HEAD
// newEpisodeRecord crea el registro. Es privado: solo el hospital genera
// episodios (con su id), porque es quien lleva la cuenta del historial.
func newEpisodeRecord(id string, p *Patient, by Attender, location string, room *Room) EpisodeRecord {
	return EpisodeRecord{
		id:         id,
		at:         time.Now(),
		patient:    p,
		attendedBy: by,
=======
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
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
		location:   location,
		room:       room,
	}
}

<<<<<<< HEAD
// ID devuelve el identificador del episodio, por ejemplo "EP-001".
func (e EpisodeRecord) ID() string { return e.id }

// At devuelve el momento en que se registró el episodio.
func (e EpisodeRecord) At() time.Time { return e.at }

// Patient devuelve el paciente que tuvo el episodio.
func (e EpisodeRecord) Patient() *Patient { return e.patient }

// AttendedBy devuelve a quien atendió el episodio (doctor o camillero).
func (e EpisodeRecord) AttendedBy() Attender { return e.attendedBy }

// Location devuelve el lugar donde le dio el ataque al paciente.
func (e EpisodeRecord) Location() string { return e.location }

// Room devuelve la sala asignada, o nil si el paciente quedó en el pasillo.
func (e EpisodeRecord) Room() *Room { return e.room }

// Summary devuelve una línea legible del episodio para imprimir en main.
// No incluye la hora para que la salida del programa sea siempre la misma.
func (e EpisodeRecord) Summary() string {
	who, patient, where := "nobody", "unknown patient", "no room (hallway)"
	if e.attendedBy != nil {
		who = e.attendedBy.Name()
	}
	if e.patient != nil {
		patient = e.patient.ID() + " " + e.patient.Name()
	}
	if e.room != nil {
		where = fmt.Sprintf("room %d", e.room.Number())
	}
	return fmt.Sprintf("%s: %s at %s, attended by %s -> %s", e.id, patient, e.location, who, where)
=======
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
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
}
