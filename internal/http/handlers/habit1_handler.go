package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/auth"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/habit1"
)

// Habit1Handler traduce HTTP <-> dominio para Habit1.
type Habit1Handler struct {
	svc *habit1.Service
}

// NewHabit1Handler construye el handler inyectando el service.
func NewHabit1Handler(svc *habit1.Service) *Habit1Handler {
	return &Habit1Handler{svc: svc}
}

// registerHabit1Request es el body de POST /habit-1.
// Fecha se recibe en formato RFC3339 (p.ej. 2026-08-22T00:00:00Z) y las horas
// como "HH:MM" en 24 horas.
//
// A propósito no tiene userId: el dueño sale del token verificado. Si el
// cliente manda uno en el body, ShouldBindJSON lo descarta.
//
// RitualNoche y RitualDia no llevan `binding:"required"`: false es un valor
// válido y required lo rechazaría por ser el cero del tipo.
type registerHabit1Request struct {
	NumeroDeRepeticion int       `json:"numeroDeRepeticion" binding:"required"`
	Fecha              time.Time `json:"fecha" binding:"required"`
	HoraDespertar      string    `json:"horaDespertar" binding:"required"`
	HoraDormir         string    `json:"horaDormir" binding:"required"`
	HorasDormidas      string    `json:"horasDormidas" binding:"required"`
	RitualNoche        bool      `json:"ritualNoche"`
	RitualDia          bool      `json:"ritualDia"`
}

// habit1Response es la representación REST de un Habit1.
type habit1Response struct {
	ID                 string    `json:"id"`
	NumeroDeRepeticion int       `json:"numeroDeRepeticion"`
	Fecha              time.Time `json:"fecha"`
	HoraDespertar      string    `json:"horaDespertar"`
	HoraDormir         string    `json:"horaDormir"`
	HorasDormidas      string    `json:"horasDormidas"`
	RitualNoche        bool      `json:"ritualNoche"`
	RitualDia          bool      `json:"ritualDia"`
}

func newHabit1Response(h *habit1.Habit1) habit1Response {
	return habit1Response{
		ID:                 h.ID,
		NumeroDeRepeticion: h.NumeroDeRepeticion,
		Fecha:              h.Fecha,
		HoraDespertar:      h.HoraDespertar,
		HoraDormir:         h.HoraDormir,
		HorasDormidas:      h.HorasDormidas,
		RitualNoche:        h.RitualNoche,
		RitualDia:          h.RitualDia,
	}
}

// Register maneja POST /habit-1
func (h *Habit1Handler) Register(c *gin.Context) {
	// El dueño sale del token que ya verificó el middleware. En una ruta
	// protegida esto siempre está; el 401 cubre que alguien monte la ruta
	// sin el middleware.
	userID, ok := auth.UID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	var req registerHabit1Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := h.svc.RegisterHabit1(
		c.Request.Context(),
		userID,
		req.NumeroDeRepeticion,
		req.Fecha,
		req.HoraDespertar,
		req.HoraDormir,
		req.HorasDormidas,
		req.RitualNoche,
		req.RitualDia,
	)
	if err != nil {
		if errors.Is(err, habit1.ErrInvalidNumeroDeRepeticion) ||
			errors.Is(err, habit1.ErrInvalidFecha) ||
			errors.Is(err, habit1.ErrInvalidHoraDespertar) ||
			errors.Is(err, habit1.ErrInvalidHoraDormir) ||
			errors.Is(err, habit1.ErrInvalidHorasDormidas) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newHabit1Response(record))
}

// List maneja GET /habit-1 y devuelve sólo los registros del usuario
// autenticado.
func (h *Habit1Handler) List(c *gin.Context) {
	userID, ok := auth.UID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	records, err := h.svc.GetHabit1Records(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]habit1Response, 0, len(records))
	for _, record := range records {
		responses = append(responses, newHabit1Response(record))
	}
	c.JSON(http.StatusOK, gin.H{"records": responses})
}
