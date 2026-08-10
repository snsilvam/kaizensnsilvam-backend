package income

import (
	"context"
	"errors"
	"testing"
	"time"
)

// testUserID es el UID de Firebase que el handler obtendría del token.
const testUserID = "firebase-uid-123"

// fakeRepository es un doble de prueba en memoria para Repository.
// stored simula la colección de Firestore y ListByUser imita el filtro por
// igualdad de la consulta, para poder afirmar con qué dueño consulta el service.
type fakeRepository struct {
	created    *Income
	err        error
	calls      int
	stored     []*Income
	listedUser string
}

func (r *fakeRepository) Create(_ context.Context, i *Income) (*Income, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	i.ID = "generated-id"
	r.created = i
	return i, nil
}

func (r *fakeRepository) ListByUser(_ context.Context, userID string) ([]*Income, error) {
	r.calls++
	r.listedUser = userID
	if r.err != nil {
		return nil, r.err
	}
	out := make([]*Income, 0, len(r.stored))
	for _, i := range r.stored {
		if i.UserID == userID {
			out = append(out, i)
		}
	}
	return out, nil
}

func TestRegisterIncome_Success(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	got, err := svc.RegisterIncome(context.Background(), testUserID, "Salario", 250000, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "generated-id" {
		t.Errorf("ID = %q, want %q", got.ID, "generated-id")
	}
	if got.UserID != testUserID {
		t.Errorf("UserID = %q, want %q", got.UserID, testUserID)
	}
	if got.Name != "Salario" {
		t.Errorf("Name = %q, want %q", got.Name, "Salario")
	}
	if got.Amount != 250000 {
		t.Errorf("Amount = %d, want %d", got.Amount, 250000)
	}
	if !got.Date.Equal(date) {
		t.Errorf("Date = %v, want %v", got.Date, date)
	}
	if repo.calls != 1 {
		t.Errorf("repo calls = %d, want 1", repo.calls)
	}
	// El repositorio debe recibir el ingreso ya con dueño: es lo que se
	// persiste en Firestore.
	if repo.created.UserID != testUserID {
		t.Errorf("UserID persistido = %q, want %q", repo.created.UserID, testUserID)
	}
}

func TestRegisterIncome_Invalid(t *testing.T) {
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		userID  string
		income  string
		amount  int64
		date    time.Time
		wantErr error
	}{
		{"empty user id", "", "Salario", 100, date, ErrInvalidUserID},
		{"blank user id", "   ", "Salario", 100, date, ErrInvalidUserID},
		{"empty name", testUserID, "", 100, date, ErrInvalidName},
		{"blank name", testUserID, "   ", 100, date, ErrInvalidName},
		{"zero amount", testUserID, "Salario", 0, date, ErrInvalidAmount},
		{"negative amount", testUserID, "Salario", -1, date, ErrInvalidAmount},
		{"zero date", testUserID, "Salario", 100, time.Time{}, ErrInvalidDate},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepository{}
			svc := NewService(repo)

			_, err := svc.RegisterIncome(context.Background(), tc.userID, tc.income, tc.amount, tc.date)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if repo.calls != 0 {
				t.Errorf("repo calls = %d, want 0 (no debe persistir si es inválido)", repo.calls)
			}
		})
	}
}

// TestRegisterIncome_SeparatesOwners comprueba que cada ingreso queda con el
// dueño que le corresponde y no se comparte estado entre llamadas.
func TestRegisterIncome_SeparatesOwners(t *testing.T) {
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	repoA := &fakeRepository{}
	a, err := NewService(repoA).RegisterIncome(context.Background(), "uid-a", "Salario", 100, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repoB := &fakeRepository{}
	b, err := NewService(repoB).RegisterIncome(context.Background(), "uid-b", "Salario", 100, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.UserID != "uid-a" {
		t.Errorf("UserID = %q, want uid-a", a.UserID)
	}
	if b.UserID != "uid-b" {
		t.Errorf("UserID = %q, want uid-b", b.UserID)
	}
}

func TestRegisterIncome_RepositoryError(t *testing.T) {
	wantErr := errors.New("firestore down")
	repo := &fakeRepository{err: wantErr}
	svc := NewService(repo)
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	got, err := svc.RegisterIncome(context.Background(), testUserID, "Salario", 100, date)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if got != nil {
		t.Errorf("income = %v, want nil", got)
	}
}

// incomesOfThreeOwners: dos ingresos de A, uno de B y uno antiguo sin dueño.
func incomesOfThreeOwners() []*Income {
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	return []*Income{
		{ID: "a1", UserID: "uid-a", Name: "Salario A", Amount: 100, Date: date},
		{ID: "b1", UserID: "uid-b", Name: "Salario B", Amount: 200, Date: date},
		{ID: "a2", UserID: "uid-a", Name: "Bono A", Amount: 300, Date: date},
		{ID: "legacy", UserID: "", Name: "Ingreso viejo", Amount: 400, Date: date},
	}
}

// TestGetIncomes_OnlyReturnsOwnRecords: el usuario A no ve lo de B ni los
// documentos antiguos sin dueño.
func TestGetIncomes_OnlyReturnsOwnRecords(t *testing.T) {
	repo := &fakeRepository{stored: incomesOfThreeOwners()}
	svc := NewService(repo)

	got, err := svc.GetIncomes(context.Background(), "uid-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2 (sólo los ingresos de uid-a)", len(got))
	}
	for _, i := range got {
		if i.UserID != "uid-a" {
			t.Errorf("ingreso %q es de %q, want uid-a", i.ID, i.UserID)
		}
	}
	// El service debe consultar con el UID recibido, no con otro.
	if repo.listedUser != "uid-a" {
		t.Errorf("repo consultado con %q, want uid-a", repo.listedUser)
	}
}

// TestGetIncomes_LegacyRecordsAreNeverReturned: ningún UID recupera los
// documentos anteriores a userId.
func TestGetIncomes_LegacyRecordsAreNeverReturned(t *testing.T) {
	repo := &fakeRepository{stored: incomesOfThreeOwners()}
	svc := NewService(repo)

	for _, uid := range []string{"uid-a", "uid-b", "uid-c"} {
		got, err := svc.GetIncomes(context.Background(), uid)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, i := range got {
			if i.ID == "legacy" {
				t.Errorf("%s recibió el ingreso antiguo sin dueño", uid)
			}
		}
	}
}

// TestGetIncomes_RequiresUserID: sin dueño no se consulta nada. Si no, un UID
// vacío serviría para barrer los documentos antiguos.
func TestGetIncomes_RequiresUserID(t *testing.T) {
	for _, uid := range []string{"", "   "} {
		repo := &fakeRepository{stored: incomesOfThreeOwners()}
		svc := NewService(repo)

		got, err := svc.GetIncomes(context.Background(), uid)
		if !errors.Is(err, ErrInvalidUserID) {
			t.Fatalf("err = %v, want ErrInvalidUserID", err)
		}
		if got != nil {
			t.Errorf("incomes = %v, want nil", got)
		}
		if repo.calls != 0 {
			t.Errorf("repo calls = %d, want 0 (no debe consultar sin dueño)", repo.calls)
		}
	}
}
