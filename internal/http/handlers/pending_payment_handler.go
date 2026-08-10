package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/auth"
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
//
// A propósito no tiene userId: el dueño sale del token verificado. Si el
// cliente manda uno en el body, ShouldBindJSON lo descarta.
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
	// El dueño sale del token que ya verificó el middleware. En una ruta
	// protegida esto siempre está; el 401 cubre que alguien monte la ruta
	// sin el middleware.
	userID, ok := auth.UID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	var req registerPendingPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pp, err := h.svc.RegisterPendingPayment(c.Request.Context(), userID, req.Name, req.Amount, req.DueDate)
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
	// El dueño sale del token, no del :id de la URL. Un pago de otro usuario
	// responde 404, igual que uno inexistente.
	userID, ok := auth.UID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	pp, err := h.svc.MarkPendingPaymentAsPaid(c.Request.Context(), userID, id)
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

// GetAll maneja GET /pending-payments y devuelve sólo los pagos del usuario
// autenticado.
func (h *PendingPaymentHandler) GetAll(c *gin.Context) {
	userID, ok := auth.UID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	payments, err := h.svc.GetAllPendingPayments(c.Request.Context(), userID)
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
