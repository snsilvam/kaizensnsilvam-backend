package firestore

import (
	"context"

	"cloud.google.com/go/firestore"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/pendingpayment"
)

const pendingPaymentCollection = "pending_payments"

// PendingPaymentRepository implementa pendingpayment.Repository sobre Firestore.
type PendingPaymentRepository struct {
	client *firestore.Client
}

// NewPendingPaymentRepository construye el repositorio con el cliente de Firestore.
func NewPendingPaymentRepository(client *firestore.Client) *PendingPaymentRepository {
	return &PendingPaymentRepository{client: client}
}

// ListUnpaid devuelve los pagos pendientes con paid == false.
// Sin OrderBy para no exigir un índice compuesto; el orden lo aplica el usecase.
func (r *PendingPaymentRepository) ListUnpaid(ctx context.Context) ([]*pendingpayment.PendingPayment, error) {
	snaps, err := r.client.Collection(pendingPaymentCollection).
		Where("paid", "==", false).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}

	payments := make([]*pendingpayment.PendingPayment, 0, len(snaps))
	for _, snap := range snaps {
		var p pendingpayment.PendingPayment
		if err := snap.DataTo(&p); err != nil {
			return nil, err
		}
		p.ID = snap.Ref.ID
		payments = append(payments, &p)
	}
	return payments, nil
}
