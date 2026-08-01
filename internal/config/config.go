package config

import "os"

type Config struct {
	Port      string
	Env       string
	ProjectID string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		Env:       getEnv("ENV", "local"),
		ProjectID: getEnv("GOOGLE_CLOUD_PROJECT", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
}
