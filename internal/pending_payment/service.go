package pending_payment

import (
	"context"
	"time"
)

// Service contiene la lógica de aplicación de PendingPayment (capa usecase).
type Service struct {
	repo Repository
}

// NewService construye el service inyectando el repositorio.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// RegisterPendingPayment valida y persiste un nuevo PendingPayment.
// Siempre crea con paid = false.
func (s *Service) RegisterPendingPayment(ctx context.Context, name string, amount int64, dueDate time.Time) (*PendingPayment, error) {
	pp := &PendingPayment{Name: name, Amount: amount, DueDate: dueDate, Paid: false}
	if err := pp.Validate(); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, pp)
}

// MarkPendingPaymentAsPaid busca el pago pendiente, valida que exista,
// lo marca como pagado y persiste el cambio.
func (s *Service) MarkPendingPaymentAsPaid(ctx context.Context, id string) (*PendingPayment, error) {
	pp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pp == nil {
		return nil, ErrNotFound
	}
	pp.Paid = true
	if err := s.repo.Update(ctx, pp); err != nil {
		return nil, err
	}
	return pp, nil
}

// GetAllPendingPayments retorna todos los pagos pendientes donde paid == false.
func (s *Service) GetAllPendingPayments(ctx context.Context) ([]*PendingPayment, error) {
	return s.repo.GetAllPending(ctx)
}
