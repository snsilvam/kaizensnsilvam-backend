package user

import (
	"net/mail"
	"strings"
)

// User es la entidad de dominio.
// El ID corresponde al ID del documento en Firestore, por eso se marca
// con firestore:"-" (no se persiste como campo, vive en la ruta del doc).
type User struct {
	ID    string `json:"id" firestore:"-"`
	Name  string `json:"name" firestore:"name"`
	Email string `json:"email" firestore:"email"`
}

// Validate aplica las reglas de negocio de la entidad.
func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return ErrInvalidName
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}
