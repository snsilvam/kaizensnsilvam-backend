package habit1

import "context"

// Repository es el contrato de persistencia para Habit1.
// La implementación concreta vive en infrastructure/firestore.
type Repository interface {
	Create(ctx context.Context, h *Habit1) (*Habit1, error)

	// ListByUser devuelve los registros cuyo dueño es userID. El filtro se
	// aplica en la consulta, no en Go.
	ListByUser(ctx context.Context, userID string) ([]*Habit1, error)
}
