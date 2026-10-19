package hospital

// =====================================================================
// STUB DE CONTRATO — Persona B reemplaza/completa este archivo.
// Solo existe para que la parte de A compile y sus tests corran.
// Lo que B NO debe cambiar sin avisar: el nombre Doctor, que embeba
// Person, y la firma de la interfaz Attender (viene del enunciado, 4.2).
// =====================================================================

// Attender: cualquier miembro del personal que puede atender un episodio.
type Attender interface {
	ID() string
	Name() string
	Attend(p *Patient, location string) (EpisodeRecord, error)
}

// Doctor — TODO(B): specialty, patients under care, episodes attended,
// DiagnosePatient, AttendSleepEmergency, MyEpisodes, Attend.
type Doctor struct {
	Person
}
