package hospital

import "fmt"

// Room es una sala con camas. Con capacidad 1 (la del escenario) pasa a
// Occupied con el primer paciente, pero el modelo soporta salas más grandes.
type Room struct {
	number    int
	capacity  int
	state     RoomState
	occupants []*Patient
}

// NewRoom crea una sala vacía. La capacidad debe ser al menos 1.
func NewRoom(number, capacity int) (*Room, error) {
	if number <= 0 {
		return nil, fmt.Errorf("%w: room number %d must be positive", ErrInvalidRoom, number)
	}
	if capacity < 1 {
		return nil, fmt.Errorf("%w: capacity %d must be at least 1", ErrInvalidRoom, capacity)
	}
	return &Room{number: number, capacity: capacity, state: Available}, nil
}

func (r *Room) Number() int      { return r.number }
func (r *Room) Capacity() int    { return r.capacity }
func (r *Room) State() RoomState { return r.state }

<<<<<<< HEAD
// Occupants devuelve una COPIA del slice: quien la reciba puede modificarla
// sin alterar la sala real (si devolviéramos r.occupants, se rompería el
// encapsulamiento porque el slice comparte el arreglo de fondo).
=======
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
func (r *Room) Occupants() []*Patient {
	out := make([]*Patient, len(r.occupants))
	copy(out, r.occupants)
	return out
}

<<<<<<< HEAD
// IsAvailable indica si queda al menos una cama libre.
=======
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
func (r *Room) IsAvailable() bool {
	return len(r.occupants) < r.capacity
}

// Occupy acuesta a p en la sala. Solo cambia la sala; el estado del
// paciente lo actualiza el hospital en AssignRoom.
func (r *Room) Occupy(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if r.indexOf(p) >= 0 {
		return fmt.Errorf("%w: %s in room %d", ErrPatientAlreadyInRoom, p.ID(), r.number)
	}
	if !r.IsAvailable() {
		return fmt.Errorf("%w: room %d (%d/%d)", ErrRoomFull, r.number, len(r.occupants), r.capacity)
	}
	r.occupants = append(r.occupants, p)
	r.updateState()
	return nil
}

// Release saca a p de la sala y la vuelve a dejar disponible.
func (r *Room) Release(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	i := r.indexOf(p)
	if i < 0 {
		return fmt.Errorf("%w: %s, room %d", ErrPatientNotInRoom, p.ID(), r.number)
	}
	r.occupants = append(r.occupants[:i], r.occupants[i+1:]...)
	r.updateState()
	return nil
}

func (r *Room) indexOf(p *Patient) int {
	for i, o := range r.occupants {
		if o == p {
			return i
		}
	}
	return -1
}

<<<<<<< HEAD
// updateState mantiene el estado coherente con la ocupación real,
// así nunca puede quedar "Occupied" con camas libres o al revés.
=======
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
func (r *Room) updateState() {
	if r.IsAvailable() {
		r.state = Available
	} else {
		r.state = Occupied
	}
}

func (r *Room) String() string {
	return fmt.Sprintf("room %d (%d/%d, %s)", r.number, len(r.occupants), r.capacity, r.state)
}
