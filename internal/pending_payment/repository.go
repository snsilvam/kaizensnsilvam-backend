package pending_payment

import "context"

// Repository es el contrato de persistencia para PendingPayment.
// La implementación concreta vive en infrastructure/firestore.
type Repository interface {
	Create(ctx context.Context, pp *PendingPayment) (*PendingPayment, error)
	GetByID(ctx context.Context, id string) (*PendingPayment, error)
	Update(ctx context.Context, pp *PendingPayment) error
	GetAllPending(ctx context.Context) ([]*PendingPayment, error)
}
