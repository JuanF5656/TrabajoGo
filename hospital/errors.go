package hospital

import "errors"

<<<<<<< HEAD
// Errores centinela del paquete. Se comparan con errors.Is, así que quien
// llama puede distinguir "no hay sala" de "sala llena" sin comparar textos.
// Cuando hace falta contexto (qué paciente, qué sala) se envuelven con
// fmt.Errorf("...: %w", ErrX).
=======
>>>>>>> c91238a5490f43666b1dd45b296ebedcdf3ba81d
var (
	ErrNilPatient             = errors.New("patient is nil")
	ErrInvalidPatient         = errors.New("invalid patient data")
	ErrInvalidRoom            = errors.New("invalid room data")
	ErrPatientAlreadyAdmitted = errors.New("patient already admitted")
	ErrPatientNotAdmitted     = errors.New("patient is not admitted in this hospital")
	ErrRoomAlreadyRegistered  = errors.New("room number already registered")
	ErrAlreadyAsleep          = errors.New("patient is already asleep")
	ErrAlreadyAwake           = errors.New("patient is already awake")
	ErrNotInHallway           = errors.New("patient is not asleep in a hallway")
	ErrRoomFull               = errors.New("room is full")
	ErrPatientAlreadyInRoom   = errors.New("patient is already in this room")
	ErrPatientNotInRoom       = errors.New("patient is not in this room")
	ErrNoRoomAvailable        = errors.New("no room available")
	ErrNoStaffAvailable       = errors.New("no staff available")
	ErrInvalidStaff           = errors.New("invalid staff data")
	ErrStaffAlreadyHired      = errors.New("staff member already hired")
	ErrNotHired               = errors.New("staff member is not hired by a hospital")
)
