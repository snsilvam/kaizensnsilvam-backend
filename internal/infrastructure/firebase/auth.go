package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

// NewAuthClient crea el cliente de Firebase Auth del Admin SDK para el
// proyecto indicado.
//
// Las credenciales se resuelven con Application Default Credentials (ADC):
// en local con `gcloud auth application-default login` y en Cloud Run con la
// service account del servicio. No hay archivos JSON de service account en el
// repositorio.
func NewAuthClient(ctx context.Context, projectID string) (*auth.Client, error) {
	if projectID == "" {
		return nil, fmt.Errorf("firebase: projectID is required")
	}

	// El ProjectID es explícito porque la verificación del ID token compara el
	// claim `aud` contra él; si ADC no lo expone, VerifyIDToken falla.
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
	if err != nil {
		return nil, fmt.Errorf("firebase: new app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase: auth client: %w", err)
	}
	return client, nil
}
