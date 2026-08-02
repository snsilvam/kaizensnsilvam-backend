package income

import "errors"

var (
	// ErrInvalidName se devuelve cuando el nombre está vacío.
	ErrInvalidName = errors.New("income name is required")

	// ErrInvalidAmount se devuelve cuando el monto no es mayor que cero.
	ErrInvalidAmount = errors.New("income amount must be greater than zero")

	// ErrInvalidDate se devuelve cuando la fecha está vacía.
	ErrInvalidDate = errors.New("income date is required")
)
