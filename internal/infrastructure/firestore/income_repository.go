package firestore

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/income"
)

const incomeCollection = "incomes"

// IncomeRepository implementa income.Repository sobre Firestore.
type IncomeRepository struct {
	client *firestore.Client
}

// NewIncomeRepository construye el repositorio con el cliente de Firestore.
func NewIncomeRepository(client *firestore.Client) *IncomeRepository {
	return &IncomeRepository{client: client}
}

// Create genera un nuevo documento y devuelve el Income con su ID asignado.
func (r *IncomeRepository) Create(ctx context.Context, i *income.Income) (*income.Income, error) {
	doc := r.client.Collection(incomeCollection).NewDoc()
	i.ID = doc.ID
	if _, err := doc.Set(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}

// ListReceived devuelve los ingresos con fecha anterior a until.
func (r *IncomeRepository) ListReceived(ctx context.Context, until time.Time) ([]*income.Income, error) {
	snaps, err := r.client.Collection(incomeCollection).
		Where("date", "<", until).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}
	return toIncomes(snaps)
}

// NextFrom devuelve el ingreso más próximo con fecha igual o posterior a from.
// El filtro y el orden usan el mismo campo, así que no requiere índice compuesto.
func (r *IncomeRepository) NextFrom(ctx context.Context, from time.Time) (*income.Income, error) {
	snaps, err := r.client.Collection(incomeCollection).
		Where("date", ">=", from).
		OrderBy("date", firestore.Asc).
		Limit(1).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}
	if len(snaps) == 0 {
		return nil, nil
	}

	incomes, err := toIncomes(snaps)
	if err != nil {
		return nil, err
	}
	return incomes[0], nil
}

// toIncomes mapea documentos de Firestore a entidades de dominio.
func toIncomes(snaps []*firestore.DocumentSnapshot) ([]*income.Income, error) {
	incomes := make([]*income.Income, 0, len(snaps))
	for _, snap := range snaps {
		var i income.Income
		if err := snap.DataTo(&i); err != nil {
			return nil, err
		}
		i.ID = snap.Ref.ID
		incomes = append(incomes, &i)
	}
	return incomes, nil
}
