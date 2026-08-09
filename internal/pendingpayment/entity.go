package pendingpayment

import (
	"strings"
	"time"
)

// PendingPayment es la entidad de dominio: un pago pendiente.
// El ID corresponde al ID del documento en Firestore, por eso se marca
// con firestore:"-" (no se persiste como campo, vive en la ruta del doc).
type PendingPayment struct {
	ID      string    `json:"id" firestore:"-"`
	Name    string    `json:"name" firestore:"name"`
	Amount  int64     `json:"amount" firestore:"amount"`
	DueDate time.Time `json:"dueDate" firestore:"dueDate"`
	Paid    bool      `json:"paid" firestore:"paid"`
}

// Validate aplica las reglas de negocio de la entidad.
func (p *PendingPayment) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrInvalidName
	}
	if p.Amount <= 0 {
		return ErrInvalidAmount
	}
	if p.DueDate.IsZero() {
		return ErrInvalidDueDate
	}
	return nil
}
