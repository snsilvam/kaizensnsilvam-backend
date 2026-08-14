package dashboard

import (
	"context"
	"sort"
	"time"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/income"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/pending_payment"
)

// Service contiene la lógica de aplicación del Dashboard (capa usecase).
// No tiene colección propia: orquesta los repositorios ya existentes.
type Service struct {
	incomes  income.Reader
	payments pending_payment.Reader

	// now se inyecta para poder fijar "hoy" en los tests.
	now func() time.Time
}

// NewService construye el service inyectando los repositorios que orquesta.
func NewService(incomes income.Reader, payments pending_payment.Reader) *Service {
	return &Service{incomes: incomes, payments: payments, now: time.Now}
}

// GetDashboard arma la vista agregada del dashboard financiero para el usuario.
func (s *Service) GetDashboard(ctx context.Context, userID string) (*Dashboard, error) {
	today := startOfDay(s.now())
	tomorrow := today.AddDate(0, 0, 1)

	// Recibido = todo ingreso con fecha de hoy o anterior.
	received, err := s.incomes.ListReceivedByUser(ctx, userID, tomorrow)
	if err != nil {
		return nil, err
	}

	next, err := s.incomes.NextFromByUser(ctx, userID, tomorrow)
	if err != nil {
		return nil, err
	}

	unpaid, err := s.payments.GetPendingByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	paid, err := s.payments.GetPaidByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Caja real: de los ingresos recibidos solo descuenta la plata que ya salió,
	// es decir los pagos realizados. Un pago pendiente todavía no es un gasto.
	available := totalIncomes(received) - totalPayments(paid)

	// Proyección: lo que quedaría si hoy se pagara todo lo pendiente. Cuando un
	// pago se marca como realizado pasa de este término al anterior, así que
	// este número no se mueve: el compromiso se cumplió, no desapareció.
	afterCommitments := available - totalPayments(unpaid)

	return &Dashboard{
		AvailableToday:            available,
		AvailableAfterCommitments: afterCommitments,
		NextIncome:                newNextIncome(next, today),
		PlanStatus:                planStatus(afterCommitments),
		PendingPayments:           newPendingPayments(unpaid),
		PendingPaymentsCount:      len(unpaid),
	}, nil
}

// planStatus deriva el estado del plan del dinero que quedaría tras cubrir los
// compromisos, no de la caja de hoy: tener plata hoy no significa ir bien si
// no alcanza para lo que se debe.
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

func totalPayments(payments []*pending_payment.PendingPayment) int64 {
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
func newPendingPayments(payments []*pending_payment.PendingPayment) []PendingPayment {
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
