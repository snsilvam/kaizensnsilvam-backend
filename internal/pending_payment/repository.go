package pending_payment

import "context"

// Repository es el contrato de persistencia para PendingPayment.
// La implementación concreta vive en infrastructure/firestore.
type Repository interface {
	Create(ctx context.Context, pp *PendingPayment) (*PendingPayment, error)

	// GetByID obtiene un pago por su ID, sin filtrar por dueño: es un get de
	// un documento y el dueño no se puede filtrar en la consulta. Quien lo
	// llama debe verificar la propiedad antes de exponer el resultado.
	GetByID(ctx context.Context, id string) (*PendingPayment, error)

	Update(ctx context.Context, pp *PendingPayment) error

	// ListPendingByUser devuelve los pagos sin pagar cuyo dueño es userID.
	// El filtro se aplica en la consulta, no en Go.
	ListPendingByUser(ctx context.Context, userID string) ([]*PendingPayment, error)
}

// Reader es el contrato de lectura para PendingPayment.
// Se mantiene aparte de Repository para que cada usecase dependa solo
// de lo que necesita (p. ej. el dashboard nunca escribe pagos).
type Reader interface {
	// GetAllPending retorna todos los pagos pendientes donde paid == false.
	GetAllPending(ctx context.Context) ([]*PendingPayment, error)

	// GetPendingByUser retorna los pagos pendientes cuyo dueño es userID.
	// Son compromisos: plata que se debe, pero que todavía no ha salido.
	GetPendingByUser(ctx context.Context, userID string) ([]*PendingPayment, error)

	// GetPaidByUser retorna los pagos ya realizados (paid == true) cuyo dueño
	// es userID. Son gastos consumados: plata que ya salió de la caja.
	GetPaidByUser(ctx context.Context, userID string) ([]*PendingPayment, error)
}
