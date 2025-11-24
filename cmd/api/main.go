package main

import (
	"context"

	"github.com/abhishek622/interviewMin/internal/cache"
	"github.com/abhishek622/interviewMin/internal/config"
	"github.com/abhishek622/interviewMin/internal/database"
	"github.com/abhishek622/interviewMin/internal/handler"
	"github.com/abhishek622/interviewMin/internal/logger"
	"github.com/abhishek622/interviewMin/internal/openai"
	"github.com/abhishek622/interviewMin/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type application struct {
	DB         *pgxpool.Pool
	Redis      *redis.Client
	OpenAI     *openai.Client
	Logger     *zap.Logger
	Config     *config.Config
	Repository *repository.Repository
	Handler    *handler.Application
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

	openaiClient := openai.NewClient(cfg.OpenAIKey)

	repo := repository.NewRepository(pool)

	handlerApp := &handler.Application{
		Logger:        log,
		UserRepo:      repo.User,
		InterviewRepo: repo.Interview,
		EntryRepo:     repo.Entry,
		SourceRepo:    repo.Source,
		JwtKey:        cfg.JwtSecret,
		JwtTTL:        cfg.JwtTTL,
		OpenAI:        openaiClient,
		OpenAIModel:   cfg.OpenAIModel,
	}

	app := &application{
		DB:         pool,
		Redis:      rclient,
		OpenAI:     openaiClient,
		Logger:     log,
		Config:     cfg,
		Repository: repo,
		Handler:    handlerApp,
	}

	if err := app.serve(); err != nil {
		sugar.Fatal(err)
	}
}
