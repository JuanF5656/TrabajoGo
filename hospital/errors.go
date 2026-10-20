package hospital

import "errors"

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
)
