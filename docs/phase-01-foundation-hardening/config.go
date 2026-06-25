//go:build ignore
// Template for Phase 1 — copy to: internal/infrastructure/config/config.go
//
// Typed, validated configuration. Fails fast with a clear message instead of
// crashing deep inside the app when a required value is missing.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Env      string   // "dev" | "staging" | "production"
	CORS     []string // allowed origins
}

type ServerConfig struct {
	Address      string        // e.g. ":8080"
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	URL string // DATABASE_URL DSN
}

type JWTConfig struct {
	Secret   string
	TTL      time.Duration // access token lifetime
	Issuer   string
}

const insecureDefaultSecret = "dev-secret-change-me"

// Load reads configuration from the environment and validates it.
func Load() (*Config, error) {
	env := getenv("APP_ENV", "dev")

	cfg := &Config{
		Env: env,
		Server: ServerConfig{
			Address:      getenv("SERVER_ADDR", ":8080"),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: DatabaseConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		JWT: JWTConfig{
			Secret: getenv("JWT_SECRET", insecureDefaultSecret),
			TTL:    24 * time.Hour,
			Issuer: "cpd-hub",
		},
		CORS: splitCSV(getenv("CORS_ORIGINS", "*")),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool { return c.Env == "production" }

func (c *Config) validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.IsProduction() && c.JWT.Secret == insecureDefaultSecret {
		return fmt.Errorf("JWT_SECRET must be set to a strong value in production")
	}
	return nil
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
