package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Env  string `default:"dev"`
	Port string `default:"8080"`

	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	JwtSecret    string `envconfig:"JWT_SECRET" required:"true"`
	AesSecretKey string `envconfig:"AES_SECRET_KEY" required:"true"`

	GroqAPIKey string `envconfig:"GROQ_API_KEY" required:"true"`
	AIModel    string `envconfig:"AI_MODEL" default:"meta-llama/llama-4-maverick-17b-128e-instruct"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
