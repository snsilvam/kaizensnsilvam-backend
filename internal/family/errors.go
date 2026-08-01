package family

import "errors"

var (
	// ErrNotFound se devuelve cuando no existe una Family con el ID dado.
	ErrNotFound = errors.New("family not found")

	// ErrInvalidName se devuelve cuando el nombre está vacío.
	ErrInvalidName = errors.New("family name is required")
)
