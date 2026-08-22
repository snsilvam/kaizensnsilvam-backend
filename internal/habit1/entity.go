package habit1

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// minutesPerDay son los minutos que tiene un día. Las horas del reloj se
// manejan como minutos desde 00:00, así que nunca salen de este rango.
const minutesPerDay = 24 * 60

// ZeroTimeOfDay es el valor por defecto de una hora del reloj: el 0 del
// modelo "HH:MM".
const ZeroTimeOfDay = "00:00"

// Habit1 es la entidad de dominio: el registro diario del hábito 1.
// El ID corresponde al ID del documento en Firestore, por eso se marca
// con firestore:"-" (no se persiste como campo, vive en la ruta del doc).
//
// UserID es el dueño del registro: el UID de Firebase del usuario autenticado.
// Nunca llega en el body de la petición, lo pone el handler a partir del token
// ya verificado.
//
// HoraDespertar, HoraDormir y HorasDormidas se guardan como "HH:MM" en 24
// horas. Las dos primeras son horas del reloj; HorasDormidas es una duración
// con el mismo formato, así que "07:30" son 7 horas y 30 minutos. Un único
// formato para los tres campos evita conversiones en el frontend y en
// Firestore.
type Habit1 struct {
	ID                 string    `json:"id" firestore:"-"`
	UserID             string    `json:"userId" firestore:"userId"`
	NumeroDeRepeticion int       `json:"numeroDeRepeticion" firestore:"numeroDeRepeticion"`
	Fecha              time.Time `json:"fecha" firestore:"fecha"`
	HoraDespertar      string    `json:"horaDespertar" firestore:"horaDespertar"`
	HoraDormir         string    `json:"horaDormir" firestore:"horaDormir"`
	HorasDormidas      string    `json:"horasDormidas" firestore:"horasDormidas"`
	RitualNoche        bool      `json:"ritualNoche" firestore:"ritualNoche"`
	RitualDia          bool      `json:"ritualDia" firestore:"ritualDia"`
}

// Validate aplica las reglas de negocio de la entidad.
func (h *Habit1) Validate() error {
	if strings.TrimSpace(h.UserID) == "" {
		return ErrInvalidUserID
	}
	if h.NumeroDeRepeticion <= 0 {
		return ErrInvalidNumeroDeRepeticion
	}
	if h.Fecha.IsZero() {
		return ErrInvalidFecha
	}
	if _, err := parseMinutes(h.HoraDespertar); err != nil {
		return ErrInvalidHoraDespertar
	}
	if _, err := parseMinutes(h.HoraDormir); err != nil {
		return ErrInvalidHoraDormir
	}
	if _, err := parseMinutes(h.HorasDormidas); err != nil {
		return ErrInvalidHorasDormidas
	}
	return nil
}

// parseMinutes convierte "HH:MM" en minutos desde 00:00.
func parseMinutes(value string) (int, error) {
	hours, minutes, found := strings.Cut(value, ":")
	if !found {
		return 0, fmt.Errorf("invalid time of day %q", value)
	}

	h, err := strconv.Atoi(hours)
	if err != nil || h < 0 || h > 23 {
		return 0, fmt.Errorf("invalid hours in %q", value)
	}

	m, err := strconv.Atoi(minutes)
	if err != nil || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid minutes in %q", value)
	}

	return h*60 + m, nil
}

// formatMinutes convierte minutos desde 00:00 en "HH:MM".
func formatMinutes(minutes int) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}
