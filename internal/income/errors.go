package income

import "errors"

var (
	// ErrInvalidUserID se devuelve cuando el ingreso no tiene dueño.
	ErrInvalidUserID = errors.New("income user id is required")

	// ErrInvalidName se devuelve cuando el nombre está vacío.
	ErrInvalidName = errors.New("income name is required")

	// ErrInvalidAmount se devuelve cuando el monto no es mayor que cero.
	ErrInvalidAmount = errors.New("income amount must be greater than zero")

	// ErrInvalidDate se devuelve cuando la fecha está vacía.
	ErrInvalidDate = errors.New("income date is required")

	// ErrNotFound se devuelve cuando el ingreso no existe.
	ErrNotFound = errors.New("income not found")
)
