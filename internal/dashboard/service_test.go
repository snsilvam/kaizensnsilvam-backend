package dashboard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/income"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/pendingpayment"
)

// today es el "hoy" fijo que usan los tests.
var today = time.Date(2026, time.August, 9, 15, 30, 0, 0, time.UTC)

func date(day int) time.Time {
	return time.Date(2026, time.August, day, 0, 0, 0, 0, time.UTC)
}

// fakeIncomeReader es un doble de prueba en memoria para income.Reader.
type fakeIncomeReader struct {
	received []*income.Income
	next     *income.Income
	err      error

	gotUntil time.Time
	gotFrom  time.Time
}

func (r *fakeIncomeReader) ListReceived(_ context.Context, until time.Time) ([]*income.Income, error) {
	r.gotUntil = until
	if r.err != nil {
		return nil, r.err
	}
	return r.received, nil
}

func (r *fakeIncomeReader) NextFrom(_ context.Context, from time.Time) (*income.Income, error) {
	r.gotFrom = from
	if r.err != nil {
		return nil, r.err
	}
	return r.next, nil
}

// fakePaymentRepository es un doble de prueba en memoria para pendingpayment.Repository.
type fakePaymentRepository struct {
	unpaid []*pendingpayment.PendingPayment
	err    error
}

func (r *fakePaymentRepository) ListUnpaid(_ context.Context) ([]*pendingpayment.PendingPayment, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.unpaid, nil
}

// newTestService construye el service con "hoy" fijo.
func newTestService(incomes income.Reader, payments pendingpayment.Repository) *Service {
	svc := NewService(incomes, payments)
	svc.now = func() time.Time { return today }
	return svc
}

func TestGetDashboard_Success(t *testing.T) {
	incomes := &fakeIncomeReader{
		received: []*income.Income{
			{ID: "i1", Name: "Salario", Amount: 250000, Date: date(1)},
			{ID: "i2", Name: "Freelance", Amount: 50000, Date: date(9)},
		},
		next: &income.Income{ID: "i3", Name: "Salario", Amount: 250000, Date: date(15)},
	}
	payments := &fakePaymentRepository{
		unpaid: []*pendingpayment.PendingPayment{
			{ID: "p1", Name: "Arriendo", Amount: 100000, DueDate: date(20)},
			{ID: "p2", Name: "Internet", Amount: 20000, DueDate: date(12)},
		},
	}

	got, err := newTestService(incomes, payments).GetDashboard(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 250000 + 50000 - (100000 + 20000)
	if got.AvailableToday != 180000 {
		t.Errorf("AvailableToday = %d, want %d", got.AvailableToday, 180000)
	}
	if got.PlanStatus != PlanStatusOnTrack {
		t.Errorf("PlanStatus = %q, want %q", got.PlanStatus, PlanStatusOnTrack)
	}

	if got.NextIncome == nil {
		t.Fatal("NextIncome = nil, want el ingreso futuro más próximo")
	}
	if got.NextIncome.Name != "Salario" {
		t.Errorf("NextIncome.Name = %q, want %q", got.NextIncome.Name, "Salario")
	}
	if got.NextIncome.Amount != 250000 {
		t.Errorf("NextIncome.Amount = %d, want %d", got.NextIncome.Amount, 250000)
	}
	if !got.NextIncome.Date.Equal(date(15)) {
		t.Errorf("NextIncome.Date = %v, want %v", got.NextIncome.Date, date(15))
	}
	if got.NextIncome.DaysRemaining != 6 {
		t.Errorf("NextIncome.DaysRemaining = %d, want %d", got.NextIncome.DaysRemaining, 6)
	}

	if got.PendingPaymentsCount != 2 {
		t.Errorf("PendingPaymentsCount = %d, want %d", got.PendingPaymentsCount, 2)
	}
	if len(got.PendingPayments) != 2 {
		t.Fatalf("len(PendingPayments) = %d, want %d", len(got.PendingPayments), 2)
	}
	// Ordenados por vencimiento, no en el orden que devuelve el repositorio.
	if got.PendingPayments[0].ID != "p2" || got.PendingPayments[1].ID != "p1" {
		t.Errorf("PendingPayments order = [%s %s], want [p2 p1]",
			got.PendingPayments[0].ID, got.PendingPayments[1].ID)
	}
	if got.PendingPayments[0].Name != "Internet" || got.PendingPayments[0].Amount != 20000 {
		t.Errorf("PendingPayments[0] = %+v", got.PendingPayments[0])
	}
}

// El corte entre "recibido" y "futuro" es el inicio de mañana, de forma
// que un ingreso con fecha de hoy cuenta como recibido.
func TestGetDashboard_DayBoundary(t *testing.T) {
	incomes := &fakeIncomeReader{}
	payments := &fakePaymentRepository{}

	if _, err := newTestService(incomes, payments).GetDashboard(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := date(10)
	if !incomes.gotUntil.Equal(want) {
		t.Errorf("ListReceived until = %v, want %v", incomes.gotUntil, want)
	}
	if !incomes.gotFrom.Equal(want) {
		t.Errorf("NextFrom from = %v, want %v", incomes.gotFrom, want)
	}
}

func TestGetDashboard_NoFutureIncome(t *testing.T) {
	incomes := &fakeIncomeReader{
		received: []*income.Income{{ID: "i1", Name: "Salario", Amount: 100, Date: date(1)}},
	}
	payments := &fakePaymentRepository{}

	got, err := newTestService(incomes, payments).GetDashboard(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.NextIncome != nil {
		t.Errorf("NextIncome = %+v, want nil", got.NextIncome)
	}
}

func TestGetDashboard_PlanStatus(t *testing.T) {
	cases := []struct {
		name     string
		incomes  int64
		payments int64
		want     string
	}{
		{"positivo", 100, 40, PlanStatusOnTrack},
		{"exacto", 100, 100, PlanStatusAtRisk},
		{"negativo", 100, 140, PlanStatusAtRisk},
		{"sin datos", 0, 0, PlanStatusAtRisk},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			incomes := &fakeIncomeReader{}
			if tc.incomes > 0 {
				incomes.received = []*income.Income{{Amount: tc.incomes, Date: date(1)}}
			}
			payments := &fakePaymentRepository{}
			if tc.payments > 0 {
				payments.unpaid = []*pendingpayment.PendingPayment{{Amount: tc.payments, DueDate: date(20)}}
			}

			got, err := newTestService(incomes, payments).GetDashboard(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.AvailableToday != tc.incomes-tc.payments {
				t.Errorf("AvailableToday = %d, want %d", got.AvailableToday, tc.incomes-tc.payments)
			}
			if got.PlanStatus != tc.want {
				t.Errorf("PlanStatus = %q, want %q", got.PlanStatus, tc.want)
			}
		})
	}
}

func TestGetDashboard_RepositoryError(t *testing.T) {
	wantErr := errors.New("firestore down")

	cases := []struct {
		name     string
		incomes  *fakeIncomeReader
		payments *fakePaymentRepository
	}{
		{"incomes", &fakeIncomeReader{err: wantErr}, &fakePaymentRepository{}},
		{"payments", &fakeIncomeReader{}, &fakePaymentRepository{err: wantErr}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := newTestService(tc.incomes, tc.payments).GetDashboard(context.Background())
			if !errors.Is(err, wantErr) {
				t.Fatalf("err = %v, want %v", err, wantErr)
			}
			if got != nil {
				t.Errorf("dashboard = %v, want nil", got)
			}
		})
	}
}
