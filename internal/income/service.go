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

// DeleteIncome borra de forma permanente el ingreso del usuario autenticado.
// Es un borrado real: el documento deja de existir y no vuelve en ninguna
// consulta.
//
// Si el ingreso existe pero es de otro usuario se devuelve ErrNotFound, igual
// que si no existiera: así cambiar el id de la URL no revela qué ids son
// válidos. Un documento antiguo sin dueño tampoco pertenece a nadie, así que
// también cae en ErrNotFound.
func (s *Service) DeleteIncome(ctx context.Context, userID, id string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}

	i, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if i == nil || i.UserID != userID {
		return ErrNotFound
	}

	return s.repo.Delete(ctx, id)
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
