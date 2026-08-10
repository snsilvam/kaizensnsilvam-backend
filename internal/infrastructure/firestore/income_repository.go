package firestore

import (
	"context"
	"sort"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/income"
)

const incomeCollection = "incomes"

// IncomeRepository implementa income.Repository sobre Firestore.
type IncomeRepository struct {
	client *firestore.Client
}

// NewIncomeRepository construye el repositorio con el cliente de Firestore.
func NewIncomeRepository(client *firestore.Client) *IncomeRepository {
	return &IncomeRepository{client: client}
}

// Create genera un nuevo documento y devuelve el Income con su ID asignado.
// El dueño (UserID) se persiste como el campo `userId` del documento, por el
// tag firestore de la entidad.
func (r *IncomeRepository) Create(ctx context.Context, i *income.Income) (*income.Income, error) {
	doc := r.client.Collection(incomeCollection).NewDoc()
	i.ID = doc.ID
	if _, err := doc.Set(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}

// ListByUser devuelve los ingresos cuyo dueño es userID.
//
// El filtro va en la consulta a Firestore, no en Go: nunca se traen los
// documentos de los demás usuarios. Una condición de igualdad no matchea los
// documentos que no tienen el campo, así que los ingresos anteriores a userId
// quedan fuera sin necesidad de un filtro extra.
//
// Sin OrderBy a propósito: una igualdad más un orden por otro campo exigiría
// un índice compuesto en Firestore.
func (r *IncomeRepository) ListByUser(ctx context.Context, userID string) ([]*income.Income, error) {
	snaps, err := r.client.Collection(incomeCollection).
		Where("userId", "==", userID).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}
	return toIncomes(snaps)
}

// ListReceived devuelve los ingresos con fecha anterior a until.
func (r *IncomeRepository) ListReceived(ctx context.Context, until time.Time) ([]*income.Income, error) {
	snaps, err := r.client.Collection(incomeCollection).
		Where("date", "<", until).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}
	return toIncomes(snaps)
}

// NextFrom devuelve el ingreso más próximo con fecha igual o posterior a from.
// El filtro y el orden usan el mismo campo, así que no requiere índice compuesto.
func (r *IncomeRepository) NextFrom(ctx context.Context, from time.Time) (*income.Income, error) {
	snaps, err := r.client.Collection(incomeCollection).
		Where("date", ">=", from).
		OrderBy("date", firestore.Asc).
		Limit(1).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}
	if len(snaps) == 0 {
		return nil, nil
	}

	incomes, err := toIncomes(snaps)
	if err != nil {
		return nil, err
	}
	return incomes[0], nil
}

// ListReceivedByUser devuelve los ingresos cuyo dueño es userID y cuya fecha es
// anterior a until. El filtro por userId va en la consulta; el filtro por fecha
// se aplica en Go para evitar requerir un índice compuesto.
func (r *IncomeRepository) ListReceivedByUser(ctx context.Context, userID string, until time.Time) ([]*income.Income, error) {
	snaps, err := r.client.Collection(incomeCollection).
		Where("userId", "==", userID).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}

	incomes, err := toIncomes(snaps)
	if err != nil {
		return nil, err
	}

	var received []*income.Income
	for _, i := range incomes {
		if i.Date.Before(until) {
			received = append(received, i)
		}
	}
	return received, nil
}

// NextFromByUser devuelve el ingreso más próximo cuyo dueño es userID y cuya
// fecha es igual o posterior a from. Devuelve (nil, nil) si no hay ninguno.
// El filtro por userId va en la consulta; el filtro por fecha y la ordenación
// se aplican en Go para evitar requerir un índice compuesto.
func (r *IncomeRepository) NextFromByUser(ctx context.Context, userID string, from time.Time) (*income.Income, error) {
	snaps, err := r.client.Collection(incomeCollection).
		Where("userId", "==", userID).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}

	if len(snaps) == 0 {
		return nil, nil
	}

	incomes, err := toIncomes(snaps)
	if err != nil {
		return nil, err
	}

	var candidates []*income.Income
	for _, i := range incomes {
		if i.Date.Equal(from) || i.Date.After(from) {
			candidates = append(candidates, i)
		}
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Date.Before(candidates[j].Date)
	})

	return candidates[0], nil
}

// toIncomes mapea documentos de Firestore a entidades de dominio.
func toIncomes(snaps []*firestore.DocumentSnapshot) ([]*income.Income, error) {
	incomes := make([]*income.Income, 0, len(snaps))
	for _, snap := range snaps {
		var i income.Income
		if err := snap.DataTo(&i); err != nil {
			return nil, err
		}
		i.ID = snap.Ref.ID
		incomes = append(incomes, &i)
	}
	return incomes, nil
}
