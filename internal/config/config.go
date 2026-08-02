package config

import (
	"os"
	"strings"
)

// defaultAllowedOrigins son los orígenes de desarrollo del frontend.
// En Cloud Run se sobreescriben con CORS_ALLOWED_ORIGINS.
const defaultAllowedOrigins = "http://localhost:5173,http://localhost:3000"

type Config struct {
	Port           string
	Env            string
	ProjectID      string
	AllowedOrigins []string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		Env:            getEnv("ENV", "local"),
		ProjectID:      getEnv("GOOGLE_CLOUD_PROJECT", ""),
		AllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", defaultAllowedOrigins)),
	}
}

// splitCSV parte una lista separada por comas y descarta elementos vacíos.
func splitCSV(value string) []string {
	var out []string
	for p := range strings.SplitSeq(value, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// getEnv trata una variable vacía como ausente: el pipeline expande
// `${{ vars.X }}` a "" cuando la variable no está definida en GitHub, y ahí
// queremos el default y no un valor vacío.
func getEnv(key, defaultValue string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return defaultValue
}
