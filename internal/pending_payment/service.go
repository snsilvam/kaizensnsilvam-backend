package pending_payment

import (
	"context"
	"strings"
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

// RegisterPendingPayment valida y persiste un nuevo PendingPayment a nombre de
// userID. Siempre crea con paid = false.
//
// userID es el UID de Firebase del usuario autenticado y lo aporta el handler
// desde el token verificado, nunca el body. Sin dueño no se persiste nada.
func (s *Service) RegisterPendingPayment(ctx context.Context, userID, name string, amount int64, dueDate time.Time) (*PendingPayment, error) {
	pp := &PendingPayment{UserID: userID, Name: name, Amount: amount, DueDate: dueDate, Paid: false}
	if err := pp.Validate(); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, pp)
}

// MarkPendingPaymentAsPaid busca el pago pendiente del usuario autenticado,
// valida que exista y le pertenezca, lo marca como pagado y persiste el cambio.
//
// Si el pago existe pero es de otro usuario se devuelve ErrNotFound, igual que
// si no existiera: así cambiar el id de la URL no revela qué ids son válidos.
// Un documento antiguo sin dueño tampoco pertenece a nadie, así que también
// cae en ErrNotFound.
func (s *Service) MarkPendingPaymentAsPaid(ctx context.Context, userID, id string) (*PendingPayment, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}

	pp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pp == nil || pp.UserID != userID {
		return nil, ErrNotFound
	}

	pp.Paid = true
	if err := s.repo.Update(ctx, pp); err != nil {
		return nil, err
	}
	return pp, nil
}

// GetAllPendingPayments retorna los pagos con paid == false del usuario
// autenticado.
//
// userID lo aporta el handler desde el token verificado, nunca la query string
// ni el body. Sin dueño no se consulta nada: un userID vacío no puede usarse
// para barrer los documentos antiguos que todavía no tienen dueño.
func (s *Service) GetAllPendingPayments(ctx context.Context, userID string) ([]*PendingPayment, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}
	return s.repo.ListPendingByUser(ctx, userID)
}
