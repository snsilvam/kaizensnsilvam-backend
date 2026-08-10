package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/dashboard"
)

// DashboardHandler traduce HTTP <-> dominio para el Dashboard.
type DashboardHandler struct {
	svc *dashboard.Service
}

// NewDashboardHandler construye el handler inyectando el service.
func NewDashboardHandler(svc *dashboard.Service) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// nextIncomeResponse es la representación REST del próximo ingreso.
type nextIncomeResponse struct {
	Name          string    `json:"name"`
	Amount        int64     `json:"amount"`
	Date          time.Time `json:"date"`
	DaysRemaining int       `json:"daysRemaining"`
}

// dashboardPendingPaymentResponse es la representación REST de un pago pendiente.
type dashboardPendingPaymentResponse struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Amount  int64     `json:"amount"`
	DueDate time.Time `json:"dueDate"`
}

// dashboardResponse es la representación REST del dashboard financiero.
type dashboardResponse struct {
	AvailableToday       int64                             `json:"availableToday"`
	NextIncome           *nextIncomeResponse               `json:"nextIncome"`
	PlanStatus           string                            `json:"planStatus"`
	PendingPayments      []dashboardPendingPaymentResponse `json:"pendingPayments"`
	PendingPaymentsCount int                               `json:"pendingPaymentsCount"`
}

func newDashboardResponse(d *dashboard.Dashboard) dashboardResponse {
	payments := make([]dashboardPendingPaymentResponse, 0, len(d.PendingPayments))
	for _, p := range d.PendingPayments {
		payments = append(payments, dashboardPendingPaymentResponse{
			ID:      p.ID,
			Name:    p.Name,
			Amount:  p.Amount,
			DueDate: p.DueDate,
		})
	}

	var next *nextIncomeResponse
	if d.NextIncome != nil {
		next = &nextIncomeResponse{
			Name:          d.NextIncome.Name,
			Amount:        d.NextIncome.Amount,
			Date:          d.NextIncome.Date,
			DaysRemaining: d.NextIncome.DaysRemaining,
		}
	}

	return dashboardResponse{
		AvailableToday:       d.AvailableToday,
		NextIncome:           next,
		PlanStatus:           d.PlanStatus,
		PendingPayments:      payments,
		PendingPaymentsCount: d.PendingPaymentsCount,
	}
}

// Get maneja GET /dashboard
func (h *DashboardHandler) Get(c *gin.Context) {
	d, err := h.svc.GetDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, newDashboardResponse(d))
}
