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
	"github.com/snsilvam/kaizensnsilvam-backend/internal/income"
)

// spyIncomeRepository captura el Income que llega al repositorio para poder
// afirmar con qué dueño se habría persistido, y con qué dueño se consultó.
// ListByUser imita el filtro por igualdad de la consulta de Firestore.
type spyIncomeRepository struct {
	created    *income.Income
	stored     []*income.Income
	listedUser string
}

func (r *spyIncomeRepository) Create(_ context.Context, i *income.Income) (*income.Income, error) {
	i.ID = "generated-id"
	r.created = i
	return i, nil
}

func (r *spyIncomeRepository) ListByUser(_ context.Context, userID string) ([]*income.Income, error) {
	r.listedUser = userID
	out := make([]*income.Income, 0, len(r.stored))
	for _, i := range r.stored {
		if i.UserID == userID {
			out = append(out, i)
		}
	}
	return out, nil
}

// newIncomeEngine monta GET y POST /incomes. Si uid no está vacío se simula el
// middleware de autenticación dejando el UID en el context del request.
func newIncomeEngine(repo income.Repository, uid string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewIncomeHandler(income.NewService(repo))

	if uid != "" {
		r.Use(func(c *gin.Context) {
			c.Request = c.Request.WithContext(auth.WithUID(c.Request.Context(), uid))
			c.Next()
		})
	}
	r.GET("/incomes", h.List)
	r.POST("/incomes", h.Register)
	return r
}

func getIncomes(engine *gin.Engine, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func postIncome(engine *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/incomes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

const validIncomeBody = `{"name":"Salario","amount":250000,"date":"2026-08-01T00:00:00Z"}`

func TestRegisterIncomeUsesUIDFromContext(t *testing.T) {
	repo := &spyIncomeRepository{}

	w := postIncome(newIncomeEngine(repo, "firebase-uid-123"), validIncomeBody)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body=%q)", w.Code, http.StatusCreated, w.Body.String())
	}
	if repo.created == nil {
		t.Fatal("el repositorio no recibió ningún ingreso")
	}
	if repo.created.UserID != "firebase-uid-123" {
		t.Errorf("UserID = %q, want el UID del token", repo.created.UserID)
	}
}

// TestRegisterIncomeIgnoresUserIDFromBody es la prueba de la regla: el dueño
// sale del token, jamás del body.
func TestRegisterIncomeIgnoresUserIDFromBody(t *testing.T) {
	repo := &spyIncomeRepository{}
	body := `{"userId":"uid-del-atacante","name":"Salario","amount":250000,"date":"2026-08-01T00:00:00Z"}`

	w := postIncome(newIncomeEngine(repo, "firebase-uid-123"), body)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body=%q)", w.Code, http.StatusCreated, w.Body.String())
	}
	if repo.created.UserID != "firebase-uid-123" {
		t.Errorf("UserID = %q, want el UID del token y no el del body", repo.created.UserID)
	}
}

func TestRegisterIncomeWithoutAuthenticationIsUnauthorized(t *testing.T) {
	repo := &spyIncomeRepository{}

	w := postIncome(newIncomeEngine(repo, ""), validIncomeBody)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if repo.created != nil {
		t.Error("se persistió un ingreso sin usuario autenticado")
	}
}

// TestIncomeResponseDoesNotExposeUserID: el contrato REST no cambia en este
// paso, el DTO de respuesta sigue sin userId.
func TestIncomeResponseDoesNotExposeUserID(t *testing.T) {
	repo := &spyIncomeRepository{}

	w := postIncome(newIncomeEngine(repo, "firebase-uid-123"), validIncomeBody)

	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("respuesta no es JSON: %v", err)
	}
	if _, exists := got["userId"]; exists {
		t.Errorf("la respuesta expone userId: %v", got)
	}
}

// storedIncomes: dos ingresos de A, uno de B y uno antiguo sin dueño.
func storedIncomes() []*income.Income {
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	return []*income.Income{
		{ID: "a1", UserID: "uid-a", Name: "Salario A", Amount: 100, Date: date},
		{ID: "b1", UserID: "uid-b", Name: "Salario B", Amount: 200, Date: date},
		{ID: "a2", UserID: "uid-a", Name: "Bono A", Amount: 300, Date: date},
		{ID: "legacy", Name: "Ingreso viejo", Amount: 400, Date: date},
	}
}

// incomeIDsFromBody extrae los ids de la respuesta de GET /incomes.
func incomeIDsFromBody(t *testing.T, w *httptest.ResponseRecorder) []string {
	t.Helper()
	var body struct {
		Incomes []struct {
			ID string `json:"id"`
		} `json:"incomes"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("respuesta no es JSON: %v", err)
	}
	ids := make([]string, 0, len(body.Incomes))
	for _, i := range body.Incomes {
		ids = append(ids, i.ID)
	}
	return ids
}

// TestListIncomesReturnsOnlyOwnRecords: A recibe lo suyo, nunca lo de B ni los
// documentos antiguos sin dueño.
func TestListIncomesReturnsOnlyOwnRecords(t *testing.T) {
	repo := &spyIncomeRepository{stored: storedIncomes()}

	w := getIncomes(newIncomeEngine(repo, "uid-a"), "/incomes")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body=%q)", w.Code, http.StatusOK, w.Body.String())
	}
	ids := incomeIDsFromBody(t, w)
	if len(ids) != 2 {
		t.Fatalf("ids = %v, want los 2 ingresos de uid-a", ids)
	}
	for _, id := range ids {
		if id == "b1" {
			t.Error("uid-a recibió un ingreso de uid-b")
		}
		if id == "legacy" {
			t.Error("uid-a recibió un ingreso antiguo sin dueño")
		}
	}
}

// TestListIncomesIgnoresUserIDFromQueryString: el UID sale del context, no de
// la query string. Pedir los ingresos de otro no cambia nada.
func TestListIncomesIgnoresUserIDFromQueryString(t *testing.T) {
	repo := &spyIncomeRepository{stored: storedIncomes()}

	w := getIncomes(newIncomeEngine(repo, "uid-a"), "/incomes?userId=uid-b")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if repo.listedUser != "uid-a" {
		t.Errorf("consultado con %q, want uid-a (el UID del token)", repo.listedUser)
	}
	for _, id := range incomeIDsFromBody(t, w) {
		if id == "b1" {
			t.Error("la query string dio acceso a los ingresos de uid-b")
		}
	}
}

func TestListIncomesWithoutAuthenticationIsUnauthorized(t *testing.T) {
	repo := &spyIncomeRepository{stored: storedIncomes()}

	w := getIncomes(newIncomeEngine(repo, ""), "/incomes")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if repo.listedUser != "" {
		t.Errorf("se consultó el repositorio con %q sin usuario autenticado", repo.listedUser)
	}
}
