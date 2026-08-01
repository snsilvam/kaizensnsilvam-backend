package user

import "context"

// Repository es el contrato de persistencia para User.
// La implementación concreta vive en infrastructure/firestore.
type Repository interface {
	Create(ctx context.Context, u *User) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
}
