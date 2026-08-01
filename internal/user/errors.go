package user

import "errors"

var (
	// ErrNotFound se devuelve cuando no existe un User con el ID dado.
	ErrNotFound = errors.New("user not found")

	// ErrInvalidName se devuelve cuando el nombre está vacío.
	ErrInvalidName = errors.New("user name is required")

	// ErrInvalidEmail se devuelve cuando el email está vacío o mal formado.
	ErrInvalidEmail = errors.New("user email is invalid")
)
