package income

import (
	"context"
	"time"
)

// Repository es el contrato de persistencia para Income.
// La implementación concreta vive en infrastructure/firestore.
type Repository interface {
	Create(ctx context.Context, i *Income) (*Income, error)

	// ListByUser devuelve los ingresos cuyo dueño es userID. El filtro se
	// aplica en la consulta, no en Go.
	ListByUser(ctx context.Context, userID string) ([]*Income, error)
}

// Reader es el contrato de lectura para Income.
// Se mantiene aparte de Repository para que cada usecase dependa solo
// de lo que necesita (p. ej. el dashboard nunca escribe ingresos).
type Reader interface {
	// ListReceived devuelve los ingresos con fecha anterior a until.
	ListReceived(ctx context.Context, until time.Time) ([]*Income, error)

	// NextFrom devuelve el ingreso más próximo con fecha igual o posterior
	// a from. Devuelve (nil, nil) si no hay ninguno.
	NextFrom(ctx context.Context, from time.Time) (*Income, error)
}
