package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS([]string{"http://localhost:5173"}))
	r.POST("/families", func(c *gin.Context) { c.Status(http.StatusCreated) })
	return r
}

func do(t *testing.T, method, origin string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/families", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	newEngine().ServeHTTP(w, req)
	return w
}

func TestPreflightAllowedOrigin(t *testing.T) {
	w := do(t, http.MethodOptions, "http://localhost:5173")

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("allow-origin = %q, want the request origin", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("allow-methods vacío en el preflight")
	}
}

func TestActualRequestGetsAllowOrigin(t *testing.T) {
	w := do(t, http.MethodPost, "http://localhost:5173")

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("allow-origin = %q, want the request origin", got)
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Errorf("vary = %q, want Origin", got)
	}
}

func TestDisallowedOriginGetsNoHeader(t *testing.T) {
	w := do(t, http.MethodPost, "http://evil.example")

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("allow-origin = %q, want vacío para un origen no permitido", got)
	}
}

func TestWildcardReflectsAnyOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS([]string{"*"}))
	r.POST("/families", func(c *gin.Context) { c.Status(http.StatusCreated) })

	req := httptest.NewRequest(http.MethodPost, "/families", nil)
	req.Header.Set("Origin", "https://app.example")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Errorf("allow-origin = %q, want el origen reflejado", got)
	}
}
