package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/family"
)

const familyCollection = "families"

// FamilyRepository implementa family.Repository sobre Firestore.
type FamilyRepository struct {
	client *firestore.Client
}

// NewFamilyRepository construye el repositorio con el cliente de Firestore.
func NewFamilyRepository(client *firestore.Client) *FamilyRepository {
	return &FamilyRepository{client: client}
}

// Create genera un nuevo documento y devuelve la Family con su ID asignado.
func (r *FamilyRepository) Create(ctx context.Context, f *family.Family) (*family.Family, error) {
	doc := r.client.Collection(familyCollection).NewDoc()
	f.ID = doc.ID
	if _, err := doc.Set(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

// GetByID recupera una Family por el ID del documento.
func (r *FamilyRepository) GetByID(ctx context.Context, id string) (*family.Family, error) {
	snap, err := r.client.Collection(familyCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, family.ErrNotFound
		}
		return nil, err
	}

	var f family.Family
	if err := snap.DataTo(&f); err != nil {
		return nil, err
	}
	f.ID = snap.Ref.ID
	return &f, nil
}
