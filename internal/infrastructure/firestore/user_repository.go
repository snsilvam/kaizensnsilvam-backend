package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/user"
)

const userCollection = "users"

// UserRepository implementa user.Repository sobre Firestore.
type UserRepository struct {
	client *firestore.Client
}

// NewUserRepository construye el repositorio con el cliente de Firestore.
func NewUserRepository(client *firestore.Client) *UserRepository {
	return &UserRepository{client: client}
}

// Create genera un nuevo documento y devuelve el User con su ID asignado.
func (r *UserRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	doc := r.client.Collection(userCollection).NewDoc()
	u.ID = doc.ID
	if _, err := doc.Set(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// GetByID recupera un User por el ID del documento.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	snap, err := r.client.Collection(userCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, user.ErrNotFound
		}
		return nil, err
	}

	var u user.User
	if err := snap.DataTo(&u); err != nil {
		return nil, err
	}
	u.ID = snap.Ref.ID
	return &u, nil
}
