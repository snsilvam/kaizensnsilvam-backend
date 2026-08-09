package pendingpayment

import "context"

// Repository es el contrato de persistencia para PendingPayment.
// La implementación concreta vive en infrastructure/firestore.
type Repository interface {
	// ListUnpaid devuelve los pagos con paid == false.
	ListUnpaid(ctx context.Context) ([]*PendingPayment, error)
}
