package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/pending_payment"
)

const pendingPaymentCollection = "pending_payments"

// PendingPaymentRepository implementa pending_payment.Repository sobre Firestore.
type PendingPaymentRepository struct {
	client *firestore.Client
}

// NewPendingPaymentRepository construye el repositorio con el cliente de Firestore.
func NewPendingPaymentRepository(client *firestore.Client) *PendingPaymentRepository {
	return &PendingPaymentRepository{client: client}
}

// Create genera un nuevo documento y devuelve el PendingPayment con su ID asignado.
func (r *PendingPaymentRepository) Create(ctx context.Context, pp *pending_payment.PendingPayment) (*pending_payment.PendingPayment, error) {
	doc := r.client.Collection(pendingPaymentCollection).NewDoc()
	pp.ID = doc.ID
	if _, err := doc.Set(ctx, pp); err != nil {
		return nil, err
	}
	return pp, nil
}

// GetByID obtiene un PendingPayment por su ID.
func (r *PendingPaymentRepository) GetByID(ctx context.Context, id string) (*pending_payment.PendingPayment, error) {
	snap, err := r.client.Collection(pendingPaymentCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, pending_payment.ErrNotFound
		}
		return nil, err
	}

	var pp pending_payment.PendingPayment
	if err := snap.DataTo(&pp); err != nil {
		return nil, err
	}
	pp.ID = snap.Ref.ID
	return &pp, nil
}

// Update actualiza un PendingPayment existente.
func (r *PendingPaymentRepository) Update(ctx context.Context, pp *pending_payment.PendingPayment) error {
	_, err := r.client.Collection(pendingPaymentCollection).Doc(pp.ID).Set(ctx, pp)
	return err
}

// GetAllPending retorna todos los pagos pendientes donde paid == false.
func (r *PendingPaymentRepository) GetAllPending(ctx context.Context) ([]*pending_payment.PendingPayment, error) {
	query := r.client.Collection(pendingPaymentCollection).Where("paid", "==", false)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var payments []*pending_payment.PendingPayment
	for _, doc := range docs {
		var pp pending_payment.PendingPayment
		if err := doc.DataTo(&pp); err != nil {
			return nil, err
		}
		pp.ID = doc.Ref.ID
		payments = append(payments, &pp)
	}
	return payments, nil
}
