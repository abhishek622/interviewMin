package main

import (
	"context"

	"github.com/abhishek622/interviewMin/internal/cache"
	"github.com/abhishek622/interviewMin/internal/config"
	"github.com/abhishek622/interviewMin/internal/database"
	"github.com/abhishek622/interviewMin/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type application struct {
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Logger *zap.Logger
	Config *config.Config
}

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, _ := logger.NewLogger(cfg.Env)
	defer log.Sync()
	sugar := log.Sugar()
	sugar.Infof("config loaded, env=%s", cfg.Env)

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		sugar.Fatal(err)
	}
	defer pool.Close()

	rclient := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err := cache.Ping(ctx, rclient); err != nil {
		sugar.Fatalf("redis ping failed: %v", err)
	}

	app := &application{
		DB:     pool,
		Redis:  rclient,
		Logger: log,
		Config: cfg,
	}

	if err := app.serve(); err != nil {
		sugar.Fatal(err)
	}
}
