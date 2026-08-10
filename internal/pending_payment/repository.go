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

// Reader es el contrato de lectura para PendingPayment.
// Se mantiene aparte de Repository para que cada usecase dependa solo
// de lo que necesita (p. ej. el dashboard nunca escribe pagos).
type Reader interface {
	// GetAllPending retorna todos los pagos pendientes donde paid == false.
	GetAllPending(ctx context.Context) ([]*PendingPayment, error)
}
