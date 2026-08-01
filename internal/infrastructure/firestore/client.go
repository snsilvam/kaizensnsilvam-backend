package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

// NewClient crea un cliente de Firestore para el proyecto indicado.
// Las credenciales se toman de GOOGLE_APPLICATION_CREDENTIALS o del
// entorno de ejecución (Cloud Run, GCE, etc.).
func NewClient(ctx context.Context, projectID string) (*firestore.Client, error) {
	if projectID == "" {
		return nil, fmt.Errorf("firestore: projectID is required")
	}
	return firestore.NewClient(ctx, projectID)
}
