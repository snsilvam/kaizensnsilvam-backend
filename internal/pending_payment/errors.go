package pending_payment

import "errors"

var (
	// ErrInvalidUserID se devuelve cuando el pago pendiente no tiene dueño.
	ErrInvalidUserID = errors.New("pending payment user id is required")

	// ErrInvalidName se devuelve cuando el nombre está vacío.
	ErrInvalidName = errors.New("pending payment name is required")

	// ErrInvalidAmount se devuelve cuando el monto no es mayor que cero.
	ErrInvalidAmount = errors.New("pending payment amount must be greater than zero")

	// ErrInvalidDueDate se devuelve cuando la fecha está vacía.
	ErrInvalidDueDate = errors.New("pending payment due date is required")

	// ErrNotFound se devuelve cuando el pago pendiente no existe.
	ErrNotFound = errors.New("pending payment not found")
)
