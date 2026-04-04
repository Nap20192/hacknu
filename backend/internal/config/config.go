package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	Port        string
	DatabaseURL string
	SwaggerHost string

	// Logging
	LogLevel  string
	LogPretty bool
	LogDir    string

	// Auth
	AdminUser     string
	AdminPassword string

	// Aggregator pipeline
	BufferCap         int
	FlushInterval     time.Duration

	// Retention / background jobs
	TelemetryRetention  time.Duration
	EmaRecalcInterval   time.Duration
}

func Load() Config {
	c := Config{
		Port:        getEnv("PORT", "8081"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://loco:loco_secret@localhost:5432/loco_twin?sslmode=disable"),
		SwaggerHost: getEnv("SWAGGER_HOST", "localhost:8082"),

		LogLevel:  getEnv("LOG_LEVEL", "INFO"),
		LogPretty: getEnv("LOG_PRETTY", "true") == "true",
		LogDir:    getEnv("LOG_DIR", ""),

		AdminUser:     getEnv("ADMIN_USER", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "changeme"),

		BufferCap:     getInt("BUFFER_CAP", 50),
		FlushInterval: getDuration("FLUSH_INTERVAL_MS", 500*time.Millisecond),

		TelemetryRetention: getDurationStr("TELEMETRY_RETENTION", 72*time.Hour),
		EmaRecalcInterval:  getDurationStr("EMA_RECALC_INTERVAL", time.Hour),
	}
	slog.Info("config loaded", "port", c.Port, "log_level", c.LogLevel)
	return c
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// getDuration reads a value in milliseconds (e.g. FLUSH_INTERVAL_MS=500).
func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return time.Duration(n) * time.Millisecond
		}
	}
	return fallback
}

// getDurationStr reads a Go duration string (e.g. TELEMETRY_RETENTION=72h).
func getDurationStr(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
