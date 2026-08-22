package habit1

import "errors"

var (
	// ErrInvalidUserID se devuelve cuando el registro no tiene dueño.
	ErrInvalidUserID = errors.New("habit1 user id is required")

	// ErrInvalidNumeroDeRepeticion se devuelve cuando el número de repetición
	// no es mayor que cero.
	ErrInvalidNumeroDeRepeticion = errors.New("habit1 numero de repeticion must be greater than zero")

	// ErrInvalidFecha se devuelve cuando la fecha está vacía.
	ErrInvalidFecha = errors.New("habit1 fecha is required")

	// ErrInvalidHoraDespertar se devuelve cuando la hora de despertar no es una
	// hora "HH:MM" válida.
	ErrInvalidHoraDespertar = errors.New("habit1 hora despertar must be a valid HH:MM time")

	// ErrInvalidHoraDormir se devuelve cuando la hora de dormir no es una hora
	// "HH:MM" válida.
	ErrInvalidHoraDormir = errors.New("habit1 hora dormir must be a valid HH:MM time")

	// ErrInvalidHorasDormidas se devuelve cuando las horas dormidas no son una
	// duración "HH:MM" válida.
	ErrInvalidHorasDormidas = errors.New("habit1 horas dormidas must be a valid HH:MM duration")
)
