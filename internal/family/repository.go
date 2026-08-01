package family

import "context"

// Repository es el contrato de persistencia para Family.
// La implementación concreta vive en infrastructure/firestore.
type Repository interface {
	Create(ctx context.Context, f *Family) (*Family, error)
	GetByID(ctx context.Context, id string) (*Family, error)
}
