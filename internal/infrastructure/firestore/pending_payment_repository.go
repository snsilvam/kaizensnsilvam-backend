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
// El dueño (UserID) se persiste como el campo `userId` del documento, por el
// tag firestore de la entidad.
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
// Set reescribe el documento con lo que traiga la entidad, así que el userId
// que se leyó en GetByID vuelve tal cual: un documento viejo sin dueño sigue
// sin dueño, no se le asigna el usuario que hace la operación.
func (r *PendingPaymentRepository) Update(ctx context.Context, pp *pending_payment.PendingPayment) error {
	_, err := r.client.Collection(pendingPaymentCollection).Doc(pp.ID).Set(ctx, pp)
	return err
}

// ListPendingByUser retorna los pagos con paid == false cuyo dueño es userID.
//
// Los dos filtros van en la consulta a Firestore, no en Go: nunca se traen los
// documentos de los demás usuarios. Una condición de igualdad no matchea los
// documentos que no tienen el campo, así que los pagos anteriores a userId
// quedan fuera sin necesidad de un filtro extra.
//
// Son dos igualdades sobre campos distintos, que Firestore resuelve con los
// índices de campo único; no hace falta un índice compuesto.
func (r *PendingPaymentRepository) ListPendingByUser(ctx context.Context, userID string) ([]*pending_payment.PendingPayment, error) {
	query := r.client.Collection(pendingPaymentCollection).
		Where("userId", "==", userID).
		Where("paid", "==", false)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	return toPendingPayments(docs)
}

// GetAllPending retorna todos los pagos pendientes donde paid == false, sin
// filtrar por usuario. Sólo lo usa el dashboard, que todavía no está acotado
// por usuario.
func (r *PendingPaymentRepository) GetAllPending(ctx context.Context) ([]*pending_payment.PendingPayment, error) {
	query := r.client.Collection(pendingPaymentCollection).Where("paid", "==", false)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	return toPendingPayments(docs)
}

// GetPendingByUser retorna los pagos pendientes cuyo dueño es userID.
func (r *PendingPaymentRepository) GetPendingByUser(ctx context.Context, userID string) ([]*pending_payment.PendingPayment, error) {
	return r.ListPendingByUser(ctx, userID)
}

// GetPaidByUser retorna los pagos con paid == true cuyo dueño es userID.
//
// Es la consulta simétrica a ListPendingByUser y vale el mismo razonamiento:
// los dos filtros van en la consulta, y por ser dos igualdades sobre campos
// distintos los resuelven los índices de campo único, sin índice compuesto.
func (r *PendingPaymentRepository) GetPaidByUser(ctx context.Context, userID string) ([]*pending_payment.PendingPayment, error) {
	query := r.client.Collection(pendingPaymentCollection).
		Where("userId", "==", userID).
		Where("paid", "==", true)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	return toPendingPayments(docs)
}

// toPendingPayments mapea documentos de Firestore a entidades de dominio.
func toPendingPayments(docs []*firestore.DocumentSnapshot) ([]*pending_payment.PendingPayment, error) {
	payments := make([]*pending_payment.PendingPayment, 0, len(docs))
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
