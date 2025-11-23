package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Env  string `default:"dev"`
	Port string `default:"8080"`

	DatabaseURL   string `envconfig:"DATABASE_URL" required:"true"`
	RedisAddr     string `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD"`
	RedisDB       int    `envconfig:"REDIS_DB" default:"0"`

	JwtSecret     string `envconfig:"JWT_SECRET" required:"true"`
	JwtTTLMinutes int    `envconfig:"JWT_TTL_MINUTES" default:"60"`

	OpenAIKey string `envconfig:"OPENAI_API_KEY"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
