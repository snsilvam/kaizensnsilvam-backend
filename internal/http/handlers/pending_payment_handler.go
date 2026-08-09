package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/pending_payment"
)

// PendingPaymentHandler traduce HTTP <-> dominio para PendingPayment.
type PendingPaymentHandler struct {
	svc *pending_payment.Service
}

// NewPendingPaymentHandler construye el handler inyectando el service.
func NewPendingPaymentHandler(svc *pending_payment.Service) *PendingPaymentHandler {
	return &PendingPaymentHandler{svc: svc}
}

// registerPendingPaymentRequest es el body de POST /pending-payments.
// DueDate se recibe en formato RFC3339 (p.ej. 2026-08-15T00:00:00Z).
type registerPendingPaymentRequest struct {
	Name    string    `json:"name" binding:"required"`
	Amount  int64     `json:"amount" binding:"required"`
	DueDate time.Time `json:"dueDate" binding:"required"`
}

// pendingPaymentResponse es la representación REST de un PendingPayment.
type pendingPaymentResponse struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Amount  int64     `json:"amount"`
	DueDate time.Time `json:"dueDate"`
	Paid    bool      `json:"paid"`
}

func newPendingPaymentResponse(pp *pending_payment.PendingPayment) pendingPaymentResponse {
	return pendingPaymentResponse{
		ID:      pp.ID,
		Name:    pp.Name,
		Amount:  pp.Amount,
		DueDate: pp.DueDate,
		Paid:    pp.Paid,
	}
}

// Register maneja POST /pending-payments
func (h *PendingPaymentHandler) Register(c *gin.Context) {
	var req registerPendingPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pp, err := h.svc.RegisterPendingPayment(c.Request.Context(), req.Name, req.Amount, req.DueDate)
	if err != nil {
		if errors.Is(err, pending_payment.ErrInvalidName) ||
			errors.Is(err, pending_payment.ErrInvalidAmount) ||
			errors.Is(err, pending_payment.ErrInvalidDueDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newPendingPaymentResponse(pp))
}

// MarkAsPaid maneja PATCH /pending-payments/:id/mark-as-paid
func (h *PendingPaymentHandler) MarkAsPaid(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	pp, err := h.svc.MarkPendingPaymentAsPaid(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pending_payment.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, newPendingPaymentResponse(pp))
}

// GetAll maneja GET /pending-payments
func (h *PendingPaymentHandler) GetAll(c *gin.Context) {
	payments, err := h.svc.GetAllPendingPayments(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []pendingPaymentResponse
	for _, pp := range payments {
		responses = append(responses, newPendingPaymentResponse(pp))
	}
	c.JSON(http.StatusOK, gin.H{"pending_payments": responses})
}
