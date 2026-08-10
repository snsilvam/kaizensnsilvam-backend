package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/auth"
	"github.com/snsilvam/kaizensnsilvam-backend/internal/pending_payment"
)

// spyPendingPaymentRepository es un repositorio en memoria que captura con qué
// dueño se escribió y se consultó. ListPendingByUser imita el doble filtro por
// igualdad de la consulta de Firestore.
type spyPendingPaymentRepository struct {
	created    *pending_payment.PendingPayment
	stored     map[string]*pending_payment.PendingPayment
	listedUser string
}

func (r *spyPendingPaymentRepository) Create(_ context.Context, pp *pending_payment.PendingPayment) (*pending_payment.PendingPayment, error) {
	pp.ID = "generated-id"
	r.created = pp
	return pp, nil
}

func (r *spyPendingPaymentRepository) GetByID(_ context.Context, id string) (*pending_payment.PendingPayment, error) {
	pp, exists := r.stored[id]
	if !exists {
		return nil, pending_payment.ErrNotFound
	}
	return pp, nil
}

func (r *spyPendingPaymentRepository) Update(_ context.Context, pp *pending_payment.PendingPayment) error {
	r.stored[pp.ID] = pp
	return nil
}

func (r *spyPendingPaymentRepository) ListPendingByUser(_ context.Context, userID string) ([]*pending_payment.PendingPayment, error) {
	r.listedUser = userID
	out := make([]*pending_payment.PendingPayment, 0, len(r.stored))
	for _, pp := range r.stored {
		if !pp.Paid && pp.UserID == userID {
			out = append(out, pp)
		}
	}
	return out, nil
}

func newSpyPendingPaymentRepository() *spyPendingPaymentRepository {
	return &spyPendingPaymentRepository{stored: make(map[string]*pending_payment.PendingPayment)}
}

// newPendingPaymentEngine monta las rutas de pagos pendientes. Si uid no está
// vacío se simula el middleware de autenticación dejando el UID en el context
// del request.
func newPendingPaymentEngine(repo pending_payment.Repository, uid string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewPendingPaymentHandler(pending_payment.NewService(repo))

	if uid != "" {
		r.Use(func(c *gin.Context) {
			c.Request = c.Request.WithContext(auth.WithUID(c.Request.Context(), uid))
			c.Next()
		})
	}
	r.GET("/pending-payments", h.GetAll)
	r.POST("/pending-payments", h.Register)
	r.PATCH("/pending-payments/:id/mark-as-paid", h.MarkAsPaid)
	return r
}

func doPendingPaymentRequest(engine *gin.Engine, method, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func postPendingPayment(engine *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/pending-payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

const validPendingPaymentBody = `{"name":"Internet","amount":90000,"dueDate":"2026-08-15T00:00:00Z"}`

func TestRegisterPendingPaymentUsesUIDFromContext(t *testing.T) {
	repo := newSpyPendingPaymentRepository()

	w := postPendingPayment(newPendingPaymentEngine(repo, "firebase-uid-123"), validPendingPaymentBody)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body=%q)", w.Code, http.StatusCreated, w.Body.String())
	}
	if repo.created == nil {
		t.Fatal("el repositorio no recibió ningún pago pendiente")
	}
	if repo.created.UserID != "firebase-uid-123" {
		t.Errorf("UserID = %q, want el UID del token", repo.created.UserID)
	}
}

// TestRegisterPendingPaymentIgnoresUserIDFromBody es la prueba de la regla: el
// dueño sale del token, jamás del body.
func TestRegisterPendingPaymentIgnoresUserIDFromBody(t *testing.T) {
	repo := newSpyPendingPaymentRepository()
	body := `{"userId":"uid-del-atacante","name":"Internet","amount":90000,"dueDate":"2026-08-15T00:00:00Z"}`

	w := postPendingPayment(newPendingPaymentEngine(repo, "firebase-uid-123"), body)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body=%q)", w.Code, http.StatusCreated, w.Body.String())
	}
	if repo.created.UserID != "firebase-uid-123" {
		t.Errorf("UserID = %q, want el UID del token y no el del body", repo.created.UserID)
	}
}

func TestRegisterPendingPaymentWithoutAuthenticationIsUnauthorized(t *testing.T) {
	repo := newSpyPendingPaymentRepository()

	w := postPendingPayment(newPendingPaymentEngine(repo, ""), validPendingPaymentBody)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if repo.created != nil {
		t.Error("se persistió un pago pendiente sin usuario autenticado")
	}
}

// storePendingPaymentsOfThreeOwners deja dos pagos de uid-a, uno de uid-b y uno
// antiguo sin dueño.
func storePendingPaymentsOfThreeOwners(repo *spyPendingPaymentRepository) {
	due := time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)
	repo.stored["a1"] = &pending_payment.PendingPayment{ID: "a1", UserID: "uid-a", Name: "Arriendo", Amount: 100, DueDate: due}
	repo.stored["b1"] = &pending_payment.PendingPayment{ID: "b1", UserID: "uid-b", Name: "Internet", Amount: 200, DueDate: due}
	repo.stored["a2"] = &pending_payment.PendingPayment{ID: "a2", UserID: "uid-a", Name: "Luz", Amount: 300, DueDate: due}
	repo.stored["legacy"] = &pending_payment.PendingPayment{ID: "legacy", Name: "Pago viejo", Amount: 400, DueDate: due}
}

// pendingPaymentIDsFromBody extrae los ids de la respuesta de GET /pending-payments.
func pendingPaymentIDsFromBody(t *testing.T, w *httptest.ResponseRecorder) []string {
	t.Helper()
	var body struct {
		PendingPayments []struct {
			ID string `json:"id"`
		} `json:"pending_payments"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("respuesta no es JSON: %v", err)
	}
	ids := make([]string, 0, len(body.PendingPayments))
	for _, pp := range body.PendingPayments {
		ids = append(ids, pp.ID)
	}
	return ids
}

// TestGetAllPendingPaymentsReturnsOnlyOwnRecords: A recibe lo suyo, nunca lo de
// B ni los documentos antiguos sin dueño.
func TestGetAllPendingPaymentsReturnsOnlyOwnRecords(t *testing.T) {
	repo := newSpyPendingPaymentRepository()
	storePendingPaymentsOfThreeOwners(repo)

	w := doPendingPaymentRequest(newPendingPaymentEngine(repo, "uid-a"), http.MethodGet, "/pending-payments")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body=%q)", w.Code, http.StatusOK, w.Body.String())
	}
	ids := pendingPaymentIDsFromBody(t, w)
	if len(ids) != 2 {
		t.Fatalf("ids = %v, want los 2 pagos de uid-a", ids)
	}
	for _, id := range ids {
		if id == "b1" {
			t.Error("uid-a recibió un pago de uid-b")
		}
		if id == "legacy" {
			t.Error("uid-a recibió un pago antiguo sin dueño")
		}
	}
}

// TestGetAllPendingPaymentsIgnoresUserIDFromQueryString: el UID sale del
// context, no de la query string.
func TestGetAllPendingPaymentsIgnoresUserIDFromQueryString(t *testing.T) {
	repo := newSpyPendingPaymentRepository()
	storePendingPaymentsOfThreeOwners(repo)

	w := doPendingPaymentRequest(newPendingPaymentEngine(repo, "uid-a"), http.MethodGet, "/pending-payments?userId=uid-b")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if repo.listedUser != "uid-a" {
		t.Errorf("consultado con %q, want uid-a (el UID del token)", repo.listedUser)
	}
	for _, id := range pendingPaymentIDsFromBody(t, w) {
		if id == "b1" {
			t.Error("la query string dio acceso a los pagos de uid-b")
		}
	}
}

func TestGetAllPendingPaymentsWithoutAuthenticationIsUnauthorized(t *testing.T) {
	repo := newSpyPendingPaymentRepository()
	storePendingPaymentsOfThreeOwners(repo)

	w := doPendingPaymentRequest(newPendingPaymentEngine(repo, ""), http.MethodGet, "/pending-payments")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if repo.listedUser != "" {
		t.Errorf("se consultó el repositorio con %q sin usuario autenticado", repo.listedUser)
	}
}

// TestMarkAsPaidOnOtherUsersRecordIsNotFound: cambiar el :id de la URL no da
// acceso al pago de otro usuario, y responde 404 igual que un id inexistente.
func TestMarkAsPaidOnOtherUsersRecordIsNotFound(t *testing.T) {
	repo := newSpyPendingPaymentRepository()
	storePendingPaymentsOfThreeOwners(repo)

	w := doPendingPaymentRequest(newPendingPaymentEngine(repo, "uid-b"), http.MethodPatch, "/pending-payments/a1/mark-as-paid")

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d (body=%q)", w.Code, http.StatusNotFound, w.Body.String())
	}
	if repo.stored["a1"].Paid {
		t.Error("uid-b modificó el pago de uid-a")
	}
}

// TestMarkAsPaidOnLegacyRecordIsNotFound: un documento sin dueño no es de
// nadie, así que nadie puede marcarlo ni se le asigna dueño.
func TestMarkAsPaidOnLegacyRecordIsNotFound(t *testing.T) {
	repo := newSpyPendingPaymentRepository()
	storePendingPaymentsOfThreeOwners(repo)

	w := doPendingPaymentRequest(newPendingPaymentEngine(repo, "uid-a"), http.MethodPatch, "/pending-payments/legacy/mark-as-paid")

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	legacy := repo.stored["legacy"]
	if legacy.Paid {
		t.Error("se modificó un documento sin dueño")
	}
	if legacy.UserID != "" {
		t.Errorf("UserID = %q, want vacío", legacy.UserID)
	}
}

// TestMarkAsPaidOnOwnRecordSucceeds es el contraste de los dos anteriores: el
// dueño sí puede.
func TestMarkAsPaidOnOwnRecordSucceeds(t *testing.T) {
	repo := newSpyPendingPaymentRepository()
	storePendingPaymentsOfThreeOwners(repo)

	w := doPendingPaymentRequest(newPendingPaymentEngine(repo, "uid-a"), http.MethodPatch, "/pending-payments/a1/mark-as-paid")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body=%q)", w.Code, http.StatusOK, w.Body.String())
	}
	if !repo.stored["a1"].Paid {
		t.Error("el pago no quedó marcado como pagado")
	}
}

func TestMarkAsPaidWithoutAuthenticationIsUnauthorized(t *testing.T) {
	repo := newSpyPendingPaymentRepository()
	storePendingPaymentsOfThreeOwners(repo)

	w := doPendingPaymentRequest(newPendingPaymentEngine(repo, ""), http.MethodPatch, "/pending-payments/a1/mark-as-paid")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if repo.stored["a1"].Paid {
		t.Error("se modificó un pago sin usuario autenticado")
	}
}
