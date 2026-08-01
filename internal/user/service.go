package user

import "context"

// Service contiene la lógica de aplicación de User.
type Service struct {
	repo Repository
}

// NewService construye el service inyectando el repositorio.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create valida y persiste un nuevo User.
func (s *Service) Create(ctx context.Context, name, email string) (*User, error) {
	u := &User{Name: name, Email: email}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, u)
}

// GetByID recupera un User por su ID.
func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, ErrNotFound
	}
	return s.repo.GetByID(ctx, id)
}
