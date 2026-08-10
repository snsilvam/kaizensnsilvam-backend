package auth

import (
	"context"
	"testing"
)

func TestUIDPresentInContext(t *testing.T) {
	ctx := WithUID(context.Background(), "firebase-uid-abc123")

	uid, ok := UID(ctx)

	if !ok {
		t.Fatal("ok = false, want true: el contexto sí tiene UID")
	}
	if uid != "firebase-uid-abc123" {
		t.Errorf("uid = %q, want firebase-uid-abc123", uid)
	}
}

func TestUIDAbsentFromContext(t *testing.T) {
	uid, ok := UID(context.Background())

	if ok {
		t.Error("ok = true, want false: la petición no pasó por el middleware")
	}
	if uid != "" {
		t.Errorf("uid = %q, want vacío cuando no hay usuario autenticado", uid)
	}
}

func TestEmptyUIDIsNotAuthenticated(t *testing.T) {
	ctx := WithUID(context.Background(), "")

	uid, ok := UID(ctx)

	if ok {
		t.Error("ok = true, want false: un UID vacío no identifica a nadie")
	}
	if uid != "" {
		t.Errorf("uid = %q, want vacío", uid)
	}
}

// otherKey imita la clave de otro paquete para comprobar que la clave privada
// de este no colisiona con valores ajenos del contexto.
type otherKey struct{}

func TestUIDIgnoresValuesUnderOtherKeys(t *testing.T) {
	ctx := context.WithValue(context.Background(), otherKey{}, "uid-de-otro-paquete")

	if _, ok := UID(ctx); ok {
		t.Error("ok = true, want false: el valor no lo escribió WithUID")
	}
}

func TestWithUIDDoesNotMutateParent(t *testing.T) {
	parent := context.Background()

	child := WithUID(parent, "uid-123")

	if _, ok := UID(parent); ok {
		t.Error("el contexto padre quedó autenticado")
	}
	if _, ok := UID(child); !ok {
		t.Error("el contexto hijo no quedó autenticado")
	}
}

func TestWithUIDOverridesInNestedContext(t *testing.T) {
	ctx := WithUID(WithUID(context.Background(), "primero"), "segundo")

	uid, _ := UID(ctx)

	if uid != "segundo" {
		t.Errorf("uid = %q, want segundo: gana el WithUID más interno", uid)
	}
}
