package hospital

// Los estados son constantes tipadas (no strings sueltos): si alguien escribe
// mal un estado, el compilador lo detecta. Cada tipo implementa fmt.Stringer
// para imprimirse de forma legible con fmt.Println / %v.

// PatientState indica dónde y cómo está un paciente.
type PatientState int

const (
	Awake           PatientState = iota // despierto
	AsleepInHallway                     // dormido en un pasillo (hay que atenderlo)
	AsleepInBed                         // dormido en una cama asignada
)

func (s PatientState) String() string {
	switch s {
	case Awake:
		return "Awake"
	case AsleepInHallway:
		return "AsleepInHallway"
	case AsleepInBed:
		return "AsleepInBed"
	default:
		return "UnknownPatientState"
	}
}

// NarcolepsyLevel es la severidad de la narcolepsia del paciente.
type NarcolepsyLevel int

const (
	Mild NarcolepsyLevel = iota
	Moderate
	Severe
)

func (l NarcolepsyLevel) String() string {
	switch l {
	case Mild:
		return "Mild"
	case Moderate:
		return "Moderate"
	case Severe:
		return "Severe"
	default:
		return "UnknownNarcolepsyLevel"
	}
}

// isValid evita que se cree un paciente con un nivel fuera del rango,
// por ejemplo NarcolepsyLevel(7).
func (l NarcolepsyLevel) isValid() bool {
	return l >= Mild && l <= Severe
}

// RoomState indica si una sala todavía acepta pacientes.
type RoomState int

const (
	Available RoomState = iota
	Occupied
)

func (s RoomState) String() string {
	switch s {
	case Available:
		return "Available"
	case Occupied:
		return "Occupied"
	default:
		return "UnknownRoomState"
	}
}
