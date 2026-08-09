package dashboard

import (
	"context"
	"sort"
	"time"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/income"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/pendingpayment"
)

// Service contiene la lógica de aplicación del Dashboard (capa usecase).
// No tiene colección propia: orquesta los repositorios ya existentes.
type Service struct {
	incomes  income.Reader
	payments pendingpayment.Repository

	// now se inyecta para poder fijar "hoy" en los tests.
	now func() time.Time
}

// NewService construye el service inyectando los repositorios que orquesta.
func NewService(incomes income.Reader, payments pendingpayment.Repository) *Service {
	return &Service{incomes: incomes, payments: payments, now: time.Now}
}

// GetDashboard arma la vista agregada del dashboard financiero.
func (s *Service) GetDashboard(ctx context.Context) (*Dashboard, error) {
	today := startOfDay(s.now())
	tomorrow := today.AddDate(0, 0, 1)

	// Recibido = todo ingreso con fecha de hoy o anterior.
	received, err := s.incomes.ListReceived(ctx, tomorrow)
	if err != nil {
		return nil, err
	}

	next, err := s.incomes.NextFrom(ctx, tomorrow)
	if err != nil {
		return nil, err
	}

	unpaid, err := s.payments.ListUnpaid(ctx)
	if err != nil {
		return nil, err
	}

	available := totalIncomes(received) - totalPayments(unpaid)

	return &Dashboard{
		AvailableToday:       available,
		NextIncome:           newNextIncome(next, today),
		PlanStatus:           planStatus(available),
		PendingPayments:      newPendingPayments(unpaid),
		PendingPaymentsCount: len(unpaid),
	}, nil
}

// planStatus deriva el estado del plan del dinero disponible hoy.
func planStatus(available int64) string {
	if available > 0 {
		return PlanStatusOnTrack
	}
	return PlanStatusAtRisk
}

func totalIncomes(incomes []*income.Income) int64 {
	var total int64
	for _, i := range incomes {
		total += i.Amount
	}
	return total
}

func totalPayments(payments []*pendingpayment.PendingPayment) int64 {
	var total int64
	for _, p := range payments {
		total += p.Amount
	}
	return total
}

// newNextIncome mapea el próximo ingreso y calcula los días que faltan.
// Devuelve nil si no hay ingresos futuros.
func newNextIncome(i *income.Income, today time.Time) *NextIncome {
	if i == nil {
		return nil
	}
	return &NextIncome{
		Name:          i.Name,
		Amount:        i.Amount,
		Date:          i.Date,
		DaysRemaining: daysBetween(today, i.Date),
	}
}

// newPendingPayments mapea los pagos pendientes ordenados por vencimiento.
func newPendingPayments(payments []*pendingpayment.PendingPayment) []PendingPayment {
	out := make([]PendingPayment, 0, len(payments))
	for _, p := range payments {
		out = append(out, PendingPayment{
			ID:      p.ID,
			Name:    p.Name,
			Amount:  p.Amount,
			DueDate: p.DueDate,
		})
	}
	sort.Slice(out, func(a, b int) bool { return out[a].DueDate.Before(out[b].DueDate) })
	return out
}

// startOfDay normaliza un instante al inicio de su día en UTC, para que la
// comparación de fechas no dependa de la hora ni de la zona del servidor.
func startOfDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// daysBetween cuenta días completos entre dos fechas, ignorando la hora.
func daysBetween(from, to time.Time) int {
	return int(startOfDay(to).Sub(startOfDay(from)).Hours() / 24)
}
