package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/auth"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/income"
)

// IncomeHandler traduce HTTP <-> dominio para Income.
type IncomeHandler struct {
	svc *income.Service
}

// NewIncomeHandler construye el handler inyectando el service.
func NewIncomeHandler(svc *income.Service) *IncomeHandler {
	return &IncomeHandler{svc: svc}
}

// registerIncomeRequest es el body de POST /incomes.
// Date se recibe en formato RFC3339 (p.ej. 2026-08-01T00:00:00Z).
//
// A propósito no tiene userId: el dueño sale del token verificado. Si el
// cliente manda uno en el body, ShouldBindJSON lo descarta.
type registerIncomeRequest struct {
	Name   string    `json:"name" binding:"required"`
	Amount int64     `json:"amount" binding:"required"`
	Date   time.Time `json:"date" binding:"required"`
}

// incomeResponse es la representación REST de un Income.
type incomeResponse struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Amount int64     `json:"amount"`
	Date   time.Time `json:"date"`
}

func newIncomeResponse(i *income.Income) incomeResponse {
	return incomeResponse{
		ID:     i.ID,
		Name:   i.Name,
		Amount: i.Amount,
		Date:   i.Date,
	}
}

// Register maneja POST /incomes
func (h *IncomeHandler) Register(c *gin.Context) {
	// El dueño sale del token que ya verificó el middleware. En una ruta
	// protegida esto siempre está; el 401 cubre que alguien monte la ruta
	// sin el middleware.
	userID, ok := auth.UID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	var req registerIncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	i, err := h.svc.RegisterIncome(c.Request.Context(), userID, req.Name, req.Amount, req.Date)
	if err != nil {
		if errors.Is(err, income.ErrInvalidName) ||
			errors.Is(err, income.ErrInvalidAmount) ||
			errors.Is(err, income.ErrInvalidDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newIncomeResponse(i))
}

// List maneja GET /incomes y devuelve sólo los ingresos del usuario autenticado.
func (h *IncomeHandler) List(c *gin.Context) {
	userID, ok := auth.UID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	incomes, err := h.svc.GetIncomes(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]incomeResponse, 0, len(incomes))
	for _, i := range incomes {
		responses = append(responses, newIncomeResponse(i))
	}
	c.JSON(http.StatusOK, gin.H{"incomes": responses})
}
