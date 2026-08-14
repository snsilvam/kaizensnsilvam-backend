package dashboard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/income"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/pending_payment"
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

func (r *fakeIncomeReader) ListReceivedByUser(_ context.Context, _ string, until time.Time) ([]*income.Income, error) {
	r.gotUntil = until
	if r.err != nil {
		return nil, r.err
	}
	return r.received, nil
}

func (r *fakeIncomeReader) NextFromByUser(_ context.Context, _ string, from time.Time) (*income.Income, error) {
	r.gotFrom = from
	if r.err != nil {
		return nil, r.err
	}
	return r.next, nil
}

// fakePaymentRepository es un doble de prueba en memoria para pending_payment.Reader.
type fakePaymentRepository struct {
	unpaid []*pending_payment.PendingPayment
	paid   []*pending_payment.PendingPayment
	err    error
}

func (r *fakePaymentRepository) GetAllPending(_ context.Context) ([]*pending_payment.PendingPayment, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.unpaid, nil
}

func (r *fakePaymentRepository) GetPendingByUser(_ context.Context, _ string) ([]*pending_payment.PendingPayment, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.unpaid, nil
}

func (r *fakePaymentRepository) GetPaidByUser(_ context.Context, _ string) ([]*pending_payment.PendingPayment, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.paid, nil
}

// newTestService construye el service con "hoy" fijo.
func newTestService(incomes income.Reader, payments pending_payment.Reader) *Service {
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
		unpaid: []*pending_payment.PendingPayment{
			{ID: "p1", Name: "Arriendo", Amount: 100000, DueDate: date(20)},
			{ID: "p2", Name: "Internet", Amount: 20000, DueDate: date(12)},
		},
		paid: []*pending_payment.PendingPayment{
			{ID: "p3", Name: "Luz", Amount: 30000, DueDate: date(5), Paid: true},
		},
	}

	got, err := newTestService(incomes, payments).GetDashboard(context.Background(), "test-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Caja: 250000 + 50000 - 30000 (solo el pago ya realizado).
	if got.AvailableToday != 270000 {
		t.Errorf("AvailableToday = %d, want %d", got.AvailableToday, 270000)
	}
	// Proyección: 270000 - (100000 + 20000) de compromisos.
	if got.AvailableAfterCommitments != 150000 {
		t.Errorf("AvailableAfterCommitments = %d, want %d", got.AvailableAfterCommitments, 150000)
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

	if _, err := newTestService(incomes, payments).GetDashboard(context.Background(), "test-user"); err != nil {
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

	got, err := newTestService(incomes, payments).GetDashboard(context.Background(), "test-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.NextIncome != nil {
		t.Errorf("NextIncome = %+v, want nil", got.NextIncome)
	}
}

// El estado del plan sale de lo que queda tras cubrir los compromisos, no de
// la caja de hoy.
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
				payments.unpaid = []*pending_payment.PendingPayment{{Amount: tc.payments, DueDate: date(20)}}
			}

			got, err := newTestService(incomes, payments).GetDashboard(context.Background(), "test-user")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Sin pagos realizados, la caja son los ingresos íntegros.
			if got.AvailableToday != tc.incomes {
				t.Errorf("AvailableToday = %d, want %d", got.AvailableToday, tc.incomes)
			}
			if got.AvailableAfterCommitments != tc.incomes-tc.payments {
				t.Errorf("AvailableAfterCommitments = %d, want %d",
					got.AvailableAfterCommitments, tc.incomes-tc.payments)
			}
			if got.PlanStatus != tc.want {
				t.Errorf("PlanStatus = %q, want %q", got.PlanStatus, tc.want)
			}
		})
	}
}

// Registrar un compromiso no gasta plata: la caja de hoy no se toca, solo baja
// la proyección. Es la mitad del bug que este cálculo tenía invertido.
func TestGetDashboard_RegisteringPendingPaymentDoesNotSpendCash(t *testing.T) {
	incomes := func() *fakeIncomeReader {
		return &fakeIncomeReader{
			received: []*income.Income{{ID: "i1", Amount: 1000000, Date: date(1)}},
		}
	}

	before, err := newTestService(incomes(), &fakePaymentRepository{}).
		GetDashboard(context.Background(), "test-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	withCommitment := &fakePaymentRepository{
		unpaid: []*pending_payment.PendingPayment{{ID: "p1", Amount: 200000, DueDate: date(20)}},
	}
	after, err := newTestService(incomes(), withCommitment).
		GetDashboard(context.Background(), "test-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if after.AvailableToday != before.AvailableToday {
		t.Errorf("AvailableToday = %d, want %d: registrar un pendiente no mueve la caja",
			after.AvailableToday, before.AvailableToday)
	}
	if after.AvailableAfterCommitments != before.AvailableAfterCommitments-200000 {
		t.Errorf("AvailableAfterCommitments = %d, want %d",
			after.AvailableAfterCommitments, before.AvailableAfterCommitments-200000)
	}
}

// Marcar un pago como realizado resta de la caja, no suma. Y no mueve la
// proyección, porque el compromiso ya estaba contado ahí.
func TestGetDashboard_MarkingAsPaidSubtractsFromCash(t *testing.T) {
	incomes := func() *fakeIncomeReader {
		return &fakeIncomeReader{
			received: []*income.Income{{ID: "i1", Amount: 1000000, Date: date(1)}},
		}
	}
	payment := &pending_payment.PendingPayment{ID: "p1", Amount: 200000, DueDate: date(20)}

	pending, err := newTestService(incomes(), &fakePaymentRepository{
		unpaid: []*pending_payment.PendingPayment{payment},
	}).GetDashboard(context.Background(), "test-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	settled, err := newTestService(incomes(), &fakePaymentRepository{
		paid: []*pending_payment.PendingPayment{payment},
	}).GetDashboard(context.Background(), "test-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if settled.AvailableToday != pending.AvailableToday-200000 {
		t.Errorf("AvailableToday = %d, want %d: pagar resta de la caja, no suma",
			settled.AvailableToday, pending.AvailableToday-200000)
	}
	if settled.AvailableAfterCommitments != pending.AvailableAfterCommitments {
		t.Errorf("AvailableAfterCommitments = %d, want %d: el compromiso se cumplió, no desapareció",
			settled.AvailableAfterCommitments, pending.AvailableAfterCommitments)
	}
	if settled.PendingPaymentsCount != 0 {
		t.Errorf("PendingPaymentsCount = %d, want 0", settled.PendingPaymentsCount)
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
			got, err := newTestService(tc.incomes, tc.payments).GetDashboard(context.Background(), "test-user")
			if !errors.Is(err, wantErr) {
				t.Fatalf("err = %v, want %v", err, wantErr)
			}
			if got != nil {
				t.Errorf("dashboard = %v, want nil", got)
			}
		})
	}
}
