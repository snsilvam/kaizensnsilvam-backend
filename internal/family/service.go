package family

import "context"

// Service contiene la lógica de aplicación de Family.
type Service struct {
	repo Repository
}

// NewService construye el service inyectando el repositorio.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create valida y persiste una nueva Family.
func (s *Service) Create(ctx context.Context, name string) (*Family, error) {
	f := &Family{Name: name}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, f)
}

// GetByID recupera una Family por su ID.
func (s *Service) GetByID(ctx context.Context, id string) (*Family, error) {
	if id == "" {
		return nil, ErrNotFound
	}
	return s.repo.GetByID(ctx, id)
}
