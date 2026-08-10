package income

import (
	"context"
	"strings"
	"time"
)

// Service contiene la lógica de aplicación de Income (capa usecase).
type Service struct {
	repo Repository
}

// NewService construye el service inyectando el repositorio.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// RegisterIncome valida y persiste un nuevo Income a nombre de userID.
//
// userID es el UID de Firebase del usuario autenticado y lo aporta el handler
// desde el token verificado, nunca el body. Sin dueño no se persiste nada.
func (s *Service) RegisterIncome(ctx context.Context, userID, name string, amount int64, date time.Time) (*Income, error) {
	i := &Income{UserID: userID, Name: name, Amount: amount, Date: date}
	if err := i.Validate(); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, i)
}

// GetIncomes devuelve los ingresos del usuario autenticado.
//
// userID lo aporta el handler desde el token verificado, nunca la query string
// ni el body. Sin dueño no se consulta nada: un userID vacío no puede usarse
// para barrer los documentos antiguos que todavía no tienen dueño.
func (s *Service) GetIncomes(ctx context.Context, userID string) ([]*Income, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}
	return s.repo.ListByUser(ctx, userID)
}
