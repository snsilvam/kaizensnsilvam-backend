package income

import "context"

// Repository es el contrato de persistencia para Income.
// La implementación concreta vive en infrastructure/firestore.
type Repository interface {
	Create(ctx context.Context, i *Income) (*Income, error)
}
