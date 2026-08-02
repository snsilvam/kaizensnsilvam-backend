package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

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
	var req registerIncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	i, err := h.svc.RegisterIncome(c.Request.Context(), req.Name, req.Amount, req.Date)
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
