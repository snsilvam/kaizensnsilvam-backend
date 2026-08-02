package income

import (
	"context"
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

// RegisterIncome valida y persiste un nuevo Income.
func (s *Service) RegisterIncome(ctx context.Context, name string, amount int64, date time.Time) (*Income, error) {
	i := &Income{Name: name, Amount: amount, Date: date}
	if err := i.Validate(); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, i)
}
