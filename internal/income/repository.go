package income

import (
	"context"
	"time"
)

// Repository es el contrato de persistencia para Income.
// La implementación concreta vive en infrastructure/firestore.
type Repository interface {
	Create(ctx context.Context, i *Income) (*Income, error)

	// GetByID obtiene un ingreso por su ID, sin filtrar por dueño: es un get
	// de un documento y el dueño no se puede filtrar en la consulta. Quien lo
	// llama debe verificar la propiedad antes de exponer el resultado.
	GetByID(ctx context.Context, id string) (*Income, error)

	// Delete borra el documento con ese ID de forma permanente. No filtra por
	// dueño, igual que GetByID: quien lo llama debe verificar la propiedad
	// antes de borrar.
	Delete(ctx context.Context, id string) error

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

	// ListReceivedByUser devuelve los ingresos cuyo dueño es userID y cuya fecha
	// es anterior a until.
	ListReceivedByUser(ctx context.Context, userID string, until time.Time) ([]*Income, error)

	// NextFromByUser devuelve el ingreso más próximo cuyo dueño es userID y cuya
	// fecha es igual o posterior a from. Devuelve (nil, nil) si no hay ninguno.
	NextFromByUser(ctx context.Context, userID string, from time.Time) (*Income, error)
}
