package income

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeRepository es un doble de prueba en memoria para Repository.
type fakeRepository struct {
	created *Income
	err     error
	calls   int
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

func TestRegisterIncome_Success(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	got, err := svc.RegisterIncome(context.Background(), "Salario", 250000, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "generated-id" {
		t.Errorf("ID = %q, want %q", got.ID, "generated-id")
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
}

func TestRegisterIncome_Invalid(t *testing.T) {
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		income  string
		amount  int64
		date    time.Time
		wantErr error
	}{
		{"empty name", "", 100, date, ErrInvalidName},
		{"blank name", "   ", 100, date, ErrInvalidName},
		{"zero amount", "Salario", 0, date, ErrInvalidAmount},
		{"negative amount", "Salario", -1, date, ErrInvalidAmount},
		{"zero date", "Salario", 100, time.Time{}, ErrInvalidDate},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepository{}
			svc := NewService(repo)

			_, err := svc.RegisterIncome(context.Background(), tc.income, tc.amount, tc.date)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if repo.calls != 0 {
				t.Errorf("repo calls = %d, want 0 (no debe persistir si es inválido)", repo.calls)
			}
		})
	}
}

func TestRegisterIncome_RepositoryError(t *testing.T) {
	wantErr := errors.New("firestore down")
	repo := &fakeRepository{err: wantErr}
	svc := NewService(repo)
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	got, err := svc.RegisterIncome(context.Background(), "Salario", 100, date)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if got != nil {
		t.Errorf("income = %v, want nil", got)
	}
}
