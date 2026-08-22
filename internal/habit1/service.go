package habit1

import (
	"context"
	"strings"
	"time"
)

// wakeUpStep es lo que se adelanta la hora de despertar de un día al
// siguiente: cada repetición se madruga 15 minutos más.
const wakeUpStep = 15

// Service contiene la lógica de aplicación de Habit1 (capa usecase).
type Service struct {
	repo Repository
}

// NewService construye el service inyectando el repositorio.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// RegisterHabit1 valida y persiste el registro del día a nombre de userID y, a
// continuación, persiste el registro del día siguiente ya preparado.
//
// userID es el UID de Firebase del usuario autenticado y lo aporta el handler
// desde el token verificado, nunca el body. Sin dueño no se persiste nada.
//
// Devuelve el registro que envió el usuario; el del día siguiente queda
// guardado pero no se devuelve, porque el usuario lo verá en la tabla.
func (s *Service) RegisterHabit1(ctx context.Context, userID string, numeroDeRepeticion int, fecha time.Time, horaDespertar, horaDormir, horasDormidas string, ritualNoche, ritualDia bool) (*Habit1, error) {
	h := &Habit1{
		UserID:             userID,
		NumeroDeRepeticion: numeroDeRepeticion,
		Fecha:              fecha,
		HoraDespertar:      horaDespertar,
		HoraDormir:         horaDormir,
		HorasDormidas:      horasDormidas,
		RitualNoche:        ritualNoche,
		RitualDia:          ritualDia,
	}
	if err := h.Validate(); err != nil {
		return nil, err
	}

	created, err := s.repo.Create(ctx, h)
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.Create(ctx, next(created)); err != nil {
		return nil, err
	}

	return created, nil
}

// GetHabit1Records devuelve los registros del usuario autenticado.
//
// userID lo aporta el handler desde el token verificado, nunca la query string
// ni el body. Sin dueño no se consulta nada.
func (s *Service) GetHabit1Records(ctx context.Context, userID string) ([]*Habit1, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}
	return s.repo.ListByUser(ctx, userID)
}

// next prepara el registro del día siguiente a partir del que acaba de
// guardarse: un día más, una repetición más y 15 minutos antes de despertar.
// El resto de campos arrancan en su valor por defecto, porque son los que el
// usuario todavía no ha vivido.
//
// La resta de los 15 minutos da la vuelta al reloj (00:05 -> 23:50) en vez de
// salirse del rango: es una hora del día, no una cuenta atrás.
//
// h ya pasó por Validate, así que HoraDespertar es una hora válida y el error
// de parseMinutes no puede darse.
func next(h *Habit1) *Habit1 {
	wakeUp, _ := parseMinutes(h.HoraDespertar)
	earlier := ((wakeUp-wakeUpStep)%minutesPerDay + minutesPerDay) % minutesPerDay

	return &Habit1{
		UserID:             h.UserID,
		NumeroDeRepeticion: h.NumeroDeRepeticion + 1,
		Fecha:              h.Fecha.AddDate(0, 0, 1),
		HoraDespertar:      formatMinutes(earlier),
		HoraDormir:         ZeroTimeOfDay,
		HorasDormidas:      ZeroTimeOfDay,
		RitualNoche:        false,
		RitualDia:          false,
	}
}
