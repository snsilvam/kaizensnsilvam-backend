package firestore

import (
	"context"

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
