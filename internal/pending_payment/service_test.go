package pending_payment

import (
	"context"
	"testing"
	"time"
)

// mockRepository es un repositorio en memoria para tests.
type mockRepository struct {
	payments map[string]*PendingPayment
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

func (m *mockRepository) GetAllPending(ctx context.Context) ([]*PendingPayment, error) {
	pending := make([]*PendingPayment, 0, len(m.payments))
	for _, pp := range m.payments {
		if !pp.Paid {
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
		"Internet",
		90000,
		time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if pp == nil {
		t.Error("expected pending payment, got nil")
	}
	if pp.ID == "" {
		t.Error("expected ID to be set")
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
		"",
		90000,
		time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	)

	if err != ErrInvalidName {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestRegisterPendingPayment_InvalidAmount(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	_, err := svc.RegisterPendingPayment(
		context.Background(),
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
		"Internet",
		90000,
		time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	)

	// Mark as paid
	paid, err := svc.MarkPendingPaymentAsPaid(context.Background(), registered.ID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if paid == nil {
		t.Error("expected pending payment, got nil")
	}
	if !paid.Paid {
		t.Error("expected paid to be true")
	}
	if paid.ID != registered.ID {
		t.Errorf("expected same ID, got %s != %s", paid.ID, registered.ID)
	}
}

func TestMarkPendingPaymentAsPaid_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	_, err := svc.MarkPendingPaymentAsPaid(context.Background(), "non-existent-id")

	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
