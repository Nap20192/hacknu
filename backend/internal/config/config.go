package config

import (
	"log/slog"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	LogLevel    string
	LogPretty   bool
	LogDir      string
}

func Load() Config {
	c := Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/hacknu?sslmode=disable"),
		LogLevel:    getEnv("LOG_LEVEL", "INFO"),
		LogPretty:   getEnv("LOG_PRETTY", "true") == "true",
		LogDir:      getEnv("LOG_DIR", ""),
	}
	slog.Info("config loaded", "port", c.Port)
	return c
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
