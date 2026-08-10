package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fbauth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"

	"github.com/snsilvam/kaizensnsilvam-backend/internal/auth"
)

// fakeVerifier reemplaza a *fbauth.Client en los tests: verificar un ID token
// real exige red y un proyecto de Firebase, así que se simula el contrato del
// Admin SDK (devuelve *fbauth.Token o error).
type fakeVerifier struct {
	uid      string
	err      error
	gotToken string
	calls    int
}

func (f *fakeVerifier) VerifyIDToken(_ context.Context, idToken string) (*fbauth.Token, error) {
	f.calls++
	f.gotToken = idToken
	if f.err != nil {
		return nil, f.err
	}
	return &fbauth.Token{UID: f.uid}, nil
}

// newAuthEngine monta una ruta protegida que responde 200 con el UID extraído,
// para poder afirmar sobre el UID que el middleware dejó en el contexto.
// El handler lee el UID igual que lo hará cualquier handler real: desde el
// context del request, con auth.UID().
func newAuthEngine(v TokenVerifier) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", Auth(v), func(c *gin.Context) {
		uid, ok := auth.UID(c.Request.Context())
		if !ok {
			c.String(http.StatusInternalServerError, "sin uid en el contexto")
			return
		}
		c.String(http.StatusOK, uid)
	})
	return r
}

func doAuth(v TokenVerifier, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	newAuthEngine(v).ServeHTTP(w, req)
	return w
}

func TestMissingAuthorizationHeaderIsUnauthorized(t *testing.T) {
	v := &fakeVerifier{uid: "uid-123"}

	w := doAuth(v, "")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if v.calls != 0 {
		t.Errorf("calls = %d, want 0: no debe verificarse nada sin header", v.calls)
	}
}

func TestMalformedAuthorizationHeaderIsUnauthorized(t *testing.T) {
	cases := map[string]string{
		"sin esquema":        "abc.def.ghi",
		"otro esquema":       "Basic dXNlcjpwYXNz",
		"bearer sin token":   "Bearer",
		"bearer token vacío": "Bearer   ",
	}

	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			v := &fakeVerifier{uid: "uid-123"}

			w := doAuth(v, header)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
			}
			if v.calls != 0 {
				t.Errorf("calls = %d, want 0: el header ni siquiera es un bearer válido", v.calls)
			}
		})
	}
}

func TestInvalidTokenIsUnauthorized(t *testing.T) {
	v := &fakeVerifier{err: errors.New("ID token has invalid signature")}

	w := doAuth(v, "Bearer token-invalido")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if v.calls != 1 {
		t.Errorf("calls = %d, want 1", v.calls)
	}
	if strings.Contains(w.Body.String(), "signature") {
		t.Errorf("body = %q: no debe filtrar el motivo interno de la verificación", w.Body.String())
	}
}

func TestValidTokenPassesAndExposesUID(t *testing.T) {
	v := &fakeVerifier{uid: "firebase-uid-abc123"}

	w := doAuth(v, "Bearer id-token-valido")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body=%q)", w.Code, http.StatusOK, w.Body.String())
	}
	if v.gotToken != "id-token-valido" {
		t.Errorf("token verificado = %q, want el que venía tras Bearer", v.gotToken)
	}
	if got := w.Body.String(); got != "firebase-uid-abc123" {
		t.Errorf("uid = %q, want el UID del token verificado", got)
	}
}

func TestBearerSchemeIsCaseInsensitive(t *testing.T) {
	v := &fakeVerifier{uid: "uid-123"}

	w := doAuth(v, "bearer id-token-valido")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestUIDIsAbsentWithoutTheMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/open", func(c *gin.Context) {
		if _, ok := auth.UID(c.Request.Context()); ok {
			t.Error("auth.UID() devolvió ok en una ruta sin el middleware Auth")
		}
		c.Status(http.StatusOK)
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/open", nil))
}

// TestEmptyUIDInVerifiedTokenIsUnauthorized cubre el borde: el token verifica
// pero no trae UID. La ruta protegida no debe ejecutarse sin usuario.
func TestEmptyUIDInVerifiedTokenIsUnauthorized(t *testing.T) {
	v := &fakeVerifier{uid: ""}

	w := doAuth(v, "Bearer id-token-sin-uid")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
