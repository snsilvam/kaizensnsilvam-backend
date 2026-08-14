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
//
// La plata se reporta con dos números porque son dos preguntas distintas y
// mezclarlas invierte el signo del gasto:
//
//   - AvailableToday es caja real. Solo la mueve la plata que efectivamente
//     entró o salió, así que registrar un pago pendiente no la toca y marcarlo
//     como pagado la baja.
//   - AvailableAfterCommitments es la proyección: lo que quedaría si hoy se
//     pagara todo lo pendiente. Esa es la que baja al registrar un compromiso,
//     y la que no se mueve al pagarlo (el monto solo cambia de lado).
type Dashboard struct {
	AvailableToday            int64
	AvailableAfterCommitments int64
	NextIncome                *NextIncome
	PlanStatus                string
	PendingPayments           []PendingPayment
	PendingPaymentsCount      int
}
