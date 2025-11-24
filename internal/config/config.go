package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Env  string `default:"dev"`
	Port string `default:"8080"`

	DatabaseURL   string `envconfig:"DATABASE_URL" required:"true"`
	RedisAddr     string `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD"`
	RedisDB       int    `envconfig:"REDIS_DB" default:"0"`

	JwtSecret string `envconfig:"JWT_SECRET" required:"true"`
	JwtTTL    int    `envconfig:"JWT_TTL_MINUTES" default:"60"`

	OpenAIKey   string `envconfig:"OPENAI_API_KEY"`
	OpenAIModel string `envconfig:"OPENAI_MODEL" default:"gpt-3.5-turbo"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
