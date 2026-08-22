package firestore

import (
	"context"
	"sort"

	"cloud.google.com/go/firestore"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/habit1"
)

const habit1Collection = "habit1"

// Habit1Repository implementa habit1.Repository sobre Firestore.
type Habit1Repository struct {
	client *firestore.Client
}

// NewHabit1Repository construye el repositorio con el cliente de Firestore.
func NewHabit1Repository(client *firestore.Client) *Habit1Repository {
	return &Habit1Repository{client: client}
}

// Create genera un nuevo documento y devuelve el Habit1 con su ID asignado.
// El dueño (UserID) se persiste como el campo `userId` del documento, por el
// tag firestore de la entidad.
func (r *Habit1Repository) Create(ctx context.Context, h *habit1.Habit1) (*habit1.Habit1, error) {
	doc := r.client.Collection(habit1Collection).NewDoc()
	h.ID = doc.ID
	if _, err := doc.Set(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

// ListByUser devuelve los registros cuyo dueño es userID, del más antiguo al
// más reciente.
//
// El filtro va en la consulta a Firestore, no en Go: nunca se traen los
// documentos de los demás usuarios. El orden se aplica en Go porque una
// igualdad más un OrderBy por otro campo exigiría un índice compuesto.
func (r *Habit1Repository) ListByUser(ctx context.Context, userID string) ([]*habit1.Habit1, error) {
	snaps, err := r.client.Collection(habit1Collection).
		Where("userId", "==", userID).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}

	records := make([]*habit1.Habit1, 0, len(snaps))
	for _, snap := range snaps {
		var h habit1.Habit1
		if err := snap.DataTo(&h); err != nil {
			return nil, err
		}
		h.ID = snap.Ref.ID
		records = append(records, &h)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].NumeroDeRepeticion < records[j].NumeroDeRepeticion
	})
	return records, nil
}
