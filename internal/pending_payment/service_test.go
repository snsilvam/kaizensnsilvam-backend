package pending_payment

import (
	"context"
	"testing"
	"time"
)

// testUserID es el UID de Firebase que el handler obtendría del token.
const testUserID = "firebase-uid-123"

// mockRepository es un repositorio en memoria para tests.
// listedUser guarda con qué dueño se consultó, para poder afirmar que el
// service propaga el UID autenticado hasta el repositorio.
type mockRepository struct {
	payments   map[string]*PendingPayment
	listedUser string
	listCalls  int
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		payments: make(map[string]*PendingPayment),
	}
}

func (m *mockRepository) Create(ctx context.Context, pp *PendingPayment) (*PendingPayment, error) {
	pp.ID = "test-id-" + time.Now().Format("20060102150405")
	m.payments[pp.ID] = pp
	return pp, nil
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*PendingPayment, error) {
	pp, exists := m.payments[id]
	if !exists {
		return nil, ErrNotFound
	}
	return pp, nil
}

func (m *mockRepository) Update(ctx context.Context, pp *PendingPayment) error {
	if _, exists := m.payments[pp.ID]; !exists {
		return ErrNotFound
	}
	m.payments[pp.ID] = pp
	return nil
}

// ListPendingByUser imita el doble filtro por igualdad de la consulta de
// Firestore: dueño y paid == false.
func (m *mockRepository) ListPendingByUser(ctx context.Context, userID string) ([]*PendingPayment, error) {
	m.listCalls++
	m.listedUser = userID
	pending := make([]*PendingPayment, 0, len(m.payments))
	for _, pp := range m.payments {
		if !pp.Paid && pp.UserID == userID {
			pending = append(pending, pp)
		}
	}
	return pending, nil
}

func TestRegisterPendingPayment(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	pp, err := svc.RegisterPendingPayment(
		context.Background(),
		testUserID,
		"Internet",
		90000,
		time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if pp == nil {
		t.Fatal("expected pending payment, got nil")
	}
	if pp.ID == "" {
		t.Error("expected ID to be set")
	}
	if pp.UserID != testUserID {
		t.Errorf("expected UserID %s, got %s", testUserID, pp.UserID)
	}
	// El repositorio debe recibir el pago ya con dueño: es lo que se persiste.
	if stored := repo.payments[pp.ID]; stored.UserID != testUserID {
		t.Errorf("expected stored UserID %s, got %s", testUserID, stored.UserID)
	}
	if pp.Paid {
		t.Error("expected paid to be false")
	}
	if pp.Name != "Internet" {
		t.Errorf("expected name Internet, got %s", pp.Name)
	}
	if pp.Amount != 90000 {
		t.Errorf("expected amount 90000, got %d", pp.Amount)
	}
}

func TestRegisterPendingPayment_InvalidName(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	_, err := svc.RegisterPendingPayment(
		context.Background(),
		testUserID,
		"",
		90000,
		time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidName {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

// TestRegisterPendingPayment_RequiresUserID: sin dueño no se persiste nada.
func TestRegisterPendingPayment_RequiresUserID(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	_, err := svc.RegisterPendingPayment(
		context.Background(),
		"",
		"Internet",
		90000,
		time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidUserID {
		t.Errorf("expected ErrInvalidUserID, got %v", err)
	}
	if len(repo.payments) != 0 {
		t.Errorf("expected nothing stored, got %d payments", len(repo.payments))
	}
}

func TestRegisterPendingPayment_InvalidAmount(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	_, err := svc.RegisterPendingPayment(
		context.Background(),
		testUserID,
		"Internet",
		0,
		time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestMarkPendingPaymentAsPaid(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Register first
	registered, _ := svc.RegisterPendingPayment(
		context.Background(),
		testUserID,
		"Internet",
		90000,
		time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	)

	// Mark as paid
	paid, err := svc.MarkPendingPaymentAsPaid(context.Background(), testUserID, registered.ID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if paid == nil {
		t.Fatal("expected pending payment, got nil")
	}
	if !paid.Paid {
		t.Error("expected paid to be true")
	}
	if paid.ID != registered.ID {
		t.Errorf("expected same ID, got %s != %s", paid.ID, registered.ID)
	}
	if paid.UserID != testUserID {
		t.Errorf("expected UserID to be preserved as %s, got %s", testUserID, paid.UserID)
	}
}

// TestMarkPendingPaymentAsPaid_OtherUsersRecordIsNotFound es la prueba central:
// cambiar el id de la URL no da acceso al pago de otro usuario, y la respuesta
// es la misma que la de un id inexistente para no revelar qué ids existen.
func TestMarkPendingPaymentAsPaid_OtherUsersRecordIsNotFound(t *testing.T) {
	repo := newMockRepository()
	repo.payments["de-uid-a"] = &PendingPayment{
		ID:      "de-uid-a",
		UserID:  "uid-a",
		Name:    "Arriendo",
		Amount:  100000,
		DueDate: time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC),
	}
	svc := NewService(repo)

	_, err := svc.MarkPendingPaymentAsPaid(context.Background(), "uid-b", "de-uid-a")

	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if repo.payments["de-uid-a"].Paid {
		t.Error("uid-b modificó el pago de uid-a")
	}
}

// TestMarkPendingPaymentAsPaid_LegacyRecordIsNotFound: un documento anterior a
// userId no es de nadie, así que nadie puede tocarlo ni se le asigna dueño.
func TestMarkPendingPaymentAsPaid_LegacyRecordIsNotFound(t *testing.T) {
	repo := newMockRepository()
	repo.payments["legacy-id"] = &PendingPayment{
		ID:      "legacy-id",
		Name:    "Arriendo",
		Amount:  100000,
		DueDate: time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC),
	}
	svc := NewService(repo)

	_, err := svc.MarkPendingPaymentAsPaid(context.Background(), testUserID, "legacy-id")

	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	stored := repo.payments["legacy-id"]
	if stored.Paid {
		t.Error("se modificó un documento sin dueño")
	}
	if stored.UserID != "" {
		t.Errorf("expected UserID to stay empty, got %s", stored.UserID)
	}
}

func TestMarkPendingPaymentAsPaid_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	_, err := svc.MarkPendingPaymentAsPaid(context.Background(), testUserID, "non-existent-id")

	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMarkPendingPaymentAsPaid_RequiresUserID(t *testing.T) {
	repo := newMockRepository()
	repo.payments["some-id"] = &PendingPayment{ID: "some-id", UserID: testUserID}
	svc := NewService(repo)

	_, err := svc.MarkPendingPaymentAsPaid(context.Background(), "", "some-id")

	if err != ErrInvalidUserID {
		t.Errorf("expected ErrInvalidUserID, got %v", err)
	}
}

// storeThreeOwners deja dos pagos de A, uno de B y uno antiguo sin dueño.
func storeThreeOwners(repo *mockRepository) {
	due := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	repo.payments["a1"] = &PendingPayment{ID: "a1", UserID: "uid-a", Name: "Arriendo", Amount: 100, DueDate: due}
	repo.payments["b1"] = &PendingPayment{ID: "b1", UserID: "uid-b", Name: "Internet", Amount: 200, DueDate: due}
	repo.payments["a2"] = &PendingPayment{ID: "a2", UserID: "uid-a", Name: "Luz", Amount: 300, DueDate: due}
	repo.payments["legacy"] = &PendingPayment{ID: "legacy", Name: "Pago viejo", Amount: 400, DueDate: due}
}

// TestGetAllPendingPayments_OnlyReturnsOwnRecords: el usuario A no ve lo de B
// ni los documentos antiguos sin dueño.
func TestGetAllPendingPayments_OnlyReturnsOwnRecords(t *testing.T) {
	repo := newMockRepository()
	storeThreeOwners(repo)
	svc := NewService(repo)

	got, err := svc.GetAllPendingPayments(context.Background(), "uid-a")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2 (sólo los pagos de uid-a)", len(got))
	}
	for _, pp := range got {
		if pp.UserID != "uid-a" {
			t.Errorf("pago %s es de %q, want uid-a", pp.ID, pp.UserID)
		}
	}
	// El service debe consultar con el UID recibido, no con otro.
	if repo.listedUser != "uid-a" {
		t.Errorf("repo consultado con %q, want uid-a", repo.listedUser)
	}
}

// TestGetAllPendingPayments_LegacyRecordsAreNeverReturned: ningún UID recupera
// los documentos anteriores a userId.
func TestGetAllPendingPayments_LegacyRecordsAreNeverReturned(t *testing.T) {
	repo := newMockRepository()
	storeThreeOwners(repo)
	svc := NewService(repo)

	for _, uid := range []string{"uid-a", "uid-b", "uid-c"} {
		got, err := svc.GetAllPendingPayments(context.Background(), uid)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		for _, pp := range got {
			if pp.ID == "legacy" {
				t.Errorf("%s recibió el pago antiguo sin dueño", uid)
			}
		}
	}
}

// TestGetAllPendingPayments_RequiresUserID: sin dueño no se consulta nada. Si
// no, un UID vacío serviría para barrer los documentos antiguos.
func TestGetAllPendingPayments_RequiresUserID(t *testing.T) {
	for _, uid := range []string{"", "   "} {
		repo := newMockRepository()
		storeThreeOwners(repo)
		svc := NewService(repo)

		got, err := svc.GetAllPendingPayments(context.Background(), uid)
		if err != ErrInvalidUserID {
			t.Fatalf("expected ErrInvalidUserID, got %v", err)
		}
		if got != nil {
			t.Errorf("payments = %v, want nil", got)
		}
		if repo.listCalls != 0 {
			t.Errorf("list calls = %d, want 0 (no debe consultar sin dueño)", repo.listCalls)
		}
	}
}
