package hospital

// Attender: cualquier miembro del personal que puede atender un episodio.
type Attender interface {
	ID() string
	Name() string
	Attend(p *Patient, location string) (EpisodeRecord, error)
}

type Doctor struct {
	Person
}
