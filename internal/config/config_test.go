package config

import (
	"slices"
	"testing"
)

func TestAllowedOriginsFromEnv(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example, https://admin.example")

	got := Load().AllowedOrigins
	want := []string{"https://app.example", "https://admin.example"}

	if !slices.Equal(got, want) {
		t.Errorf("AllowedOrigins = %q, want %q", got, want)
	}
}

// El pipeline expande `${{ vars.CORS_ALLOWED_ORIGINS }}` a "" si la variable
// no existe en GitHub; sin fallback quedaríamos sin ningún origen permitido.
func TestEmptyEnvFallsBackToDefaults(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	got := Load().AllowedOrigins

	if len(got) == 0 {
		t.Fatal("AllowedOrigins vacío, se esperaban los orígenes por defecto")
	}
	if !slices.Contains(got, "http://localhost:5173") {
		t.Errorf("AllowedOrigins = %q, want que incluya el localhost de dev", got)
	}
}

func TestPortDefault(t *testing.T) {
	t.Setenv("PORT", "")

	if got := Load().Port; got != "8080" {
		t.Errorf("Port = %q, want 8080", got)
	}
}
