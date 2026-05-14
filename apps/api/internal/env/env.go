package env

import (
	"fmt"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL string
	Port        string
	Domain      string
	LogLevel    string
}

func Load() (Config, error) {
	env := Config{
		DatabaseURL: valueOrDefault("DATABASE_URL", "postgres://postgres:postgres@db:5432/vision?sslmode=disable"),
		Port:        valueOrDefault("PORT", "4000"),
		Domain:      valueOrDefault("DOMAIN", "http://localhost:5173"),
		LogLevel:    valueOrDefault("LOG_LEVEL", "info"),
	}

	port, err := strconv.Atoi(env.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a valid TCP port")
	}
	if err := validateLogLevel(env.LogLevel); err != nil {
		return Config{}, err
	}

	return env, nil
}

func valueOrDefault(key string, fallback string) string {
	if value := envGet(key); value != "" {
		return value
	}
	return fallback
}

func validateLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error")
	}
}
