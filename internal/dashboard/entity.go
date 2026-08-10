package dashboard

import "time"

// Estados del plan. En V1 se derivan del dinero disponible hoy;
// no hay entidad Plan ni persistencia propia.
const (
	PlanStatusOnTrack = "on_track"
	PlanStatusAtRisk  = "at_risk"
)

// NextIncome es el ingreso futuro más próximo.
type NextIncome struct {
	Name          string
	Amount        int64
	Date          time.Time
	DaysRemaining int
}

// PendingPayment es un pago pendiente tal como lo muestra el dashboard.
type PendingPayment struct {
	ID      string
	Name    string
	Amount  int64
	DueDate time.Time
}

// Dashboard es la vista agregada que responde las tres preguntas del producto:
// cuánto puedo gastar hoy, cuándo cobro y si voy bien o mal.
// Es un modelo de lectura: no se persiste en ninguna colección.
type Dashboard struct {
	AvailableToday       int64
	NextIncome           *NextIncome
	PlanStatus           string
	PendingPayments      []PendingPayment
	PendingPaymentsCount int
}
