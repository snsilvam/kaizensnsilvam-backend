package income

import (
	"strings"
	"time"
)

// Income es la entidad de dominio: un ingreso registrado.
// El ID corresponde al ID del documento en Firestore, por eso se marca
// con firestore:"-" (no se persiste como campo, vive en la ruta del doc).
type Income struct {
	ID     string    `json:"id" firestore:"-"`
	Name   string    `json:"name" firestore:"name"`
	Amount int64     `json:"amount" firestore:"amount"`
	Date   time.Time `json:"date" firestore:"date"`
}

// Validate aplica las reglas de negocio de la entidad.
func (i *Income) Validate() error {
	if strings.TrimSpace(i.Name) == "" {
		return ErrInvalidName
	}
	if i.Amount <= 0 {
		return ErrInvalidAmount
	}
	if i.Date.IsZero() {
		return ErrInvalidDate
	}
	return nil
}
