package pendingpayment

import "errors"

var (
	// ErrInvalidName se devuelve cuando el nombre está vacío.
	ErrInvalidName = errors.New("pending payment name is required")

	// ErrInvalidAmount se devuelve cuando el monto no es mayor que cero.
	ErrInvalidAmount = errors.New("pending payment amount must be greater than zero")

	// ErrInvalidDueDate se devuelve cuando la fecha de vencimiento está vacía.
	ErrInvalidDueDate = errors.New("pending payment due date is required")
)
