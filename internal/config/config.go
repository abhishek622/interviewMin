package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration
type Config struct {
	Env     string `envconfig:"APP_ENV" default:"development"`
	Port    int    `envconfig:"APP_PORT" default:"8080"`
	DB      DBConfig
	Limiter RateLimiterConfig
	CORS    CORSConfig
	JWT     JWTConfig
	Crypto  CryptoConfig
	Groq    GroqConfig
}

// database configuration
type DBConfig struct {
	DSN          string        `envconfig:"DATABASE_URL" required:"true"`
	MaxOpenConns int           `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns int           `envconfig:"DB_MAX_IDLE_CONNS" default:"25"`
	MaxIdleTime  time.Duration `envconfig:"DB_MAX_IDLE_TIME" default:"15m"`
}

// rate limiting configuration
type RateLimiterConfig struct {
	RPS     float64 `envconfig:"RATE_LIMIT_RPS" default:"10"`
	Burst   int     `envconfig:"RATE_LIMIT_BURST" default:"20"`
	Enabled bool    `envconfig:"RATE_LIMIT_ENABLED" default:"true"`
}

// CORS configuration
type CORSConfig struct {
	TrustedOrigins []string `envconfig:"CORS_TRUSTED_ORIGINS" default:"http://localhost:3000,http://localhost:4173,http://localhost:5173"`
}

// JWT configuration
type JWTConfig struct {
	Secret          string        `envconfig:"JWT_SECRET" required:"true"`
	AccessTokenTTL  time.Duration `envconfig:"JWT_ACCESS_TOKEN_TTL" default:"15m"`
	RefreshTokenTTL time.Duration `envconfig:"JWT_REFRESH_TOKEN_TTL" default:"168h"` // 7 days
}

// encryption configuration
type CryptoConfig struct {
	Secret string `envconfig:"AES_SECRET_KEY" required:"true"`
}

// Groq AI configuration
type GroqConfig struct {
	APIKey  string        `envconfig:"GROQ_API_KEY" required:"true"`
	Model   string        `envconfig:"GROQ_MODEL" default:"meta-llama/llama-4-maverick-17b-128e-instruct"`
	Timeout time.Duration `envconfig:"GROQ_TIMEOUT" default:"30s"`
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

func (c *Config) Validate() error {
	// Validate environment
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
		"test":        true,
	}
	if !validEnvs[c.Env] {
		return fmt.Errorf("invalid environment: %s (must be one of: development, staging, production, test)", c.Env)
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be between 1 and 65535)", c.Port)
	}
	if c.DB.MaxOpenConns < 1 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be at least 1")
	}
	if c.DB.MaxIdleConns < 1 {
		return fmt.Errorf("DB_MAX_IDLE_CONNS must be at least 1")
	}
	if c.DB.MaxIdleConns > c.DB.MaxOpenConns {
		return fmt.Errorf("DB_MAX_IDLE_CONNS (%d) cannot exceed DB_MAX_OPEN_CONNS (%d)",
			c.DB.MaxIdleConns, c.DB.MaxOpenConns)
	}
	if c.Limiter.RPS < 0 {
		return fmt.Errorf("RATE_LIMIT_RPS must be non-negative")
	}
	if c.Limiter.Burst < 1 {
		return fmt.Errorf("RATE_LIMIT_BURST must be at least 1")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	secretLen := len(c.Crypto.Secret)
	if secretLen != 16 && secretLen != 24 && secretLen != 32 {
		return fmt.Errorf("AES_SECRET_KEY must be 16, 24, or 32 bytes (got %d)", secretLen)
	}
	if len(c.CORS.TrustedOrigins) == 0 {
		return fmt.Errorf("at least one trusted origin must be specified")
	}

	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) GetServerAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}

// GetCORSOrigins returns the list of trusted CORS origins
func (c *Config) GetCORSOrigins() []string {
	origins := make([]string, 0, len(c.CORS.TrustedOrigins))
	for _, origin := range c.CORS.TrustedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{Env=%s, Port=%d, DB.MaxOpenConns=%d, DB.MaxIdleConns=%d, "+
		"Limiter.RPS=%.2f, Limiter.Burst=%d, Limiter.Enabled=%t, CORS.Origins=%d, "+
		"JWT.AccessTokenTTL=%s, JWT.RefreshTokenTTL=%s, Groq.Model=%s}",
		c.Env, c.Port, c.DB.MaxOpenConns, c.DB.MaxIdleConns,
		c.Limiter.RPS, c.Limiter.Burst, c.Limiter.Enabled, len(c.CORS.TrustedOrigins),
		c.JWT.AccessTokenTTL, c.JWT.RefreshTokenTTL, c.Groq.Model)
}
