package main

import (
	"context"
	"time"

	"github.com/abhishek622/interviewMin/internal/auth"
	"github.com/abhishek622/interviewMin/internal/config"
	"github.com/abhishek622/interviewMin/internal/database"
	"github.com/abhishek622/interviewMin/internal/handler"
	"github.com/abhishek622/interviewMin/internal/logger"
	"github.com/abhishek622/interviewMin/internal/openai"
	"github.com/abhishek622/interviewMin/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
)

type application struct {
	DB         *pgxpool.Pool
	OpenAI     *openai.Client
	Logger     *zap.Logger
	Config     *config.Config
	Repository *repository.Repository
	Handler    *handler.Handler
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

	openaiClient := openai.NewClient(cfg.OpenAIKey)
	repo := repository.NewRepository(pool)
	tokenMaker := auth.NewJWTMaker(cfg.JwtSecret)

	handlerApp := &handler.Handler{
		Logger:         log,
		UserRepo:       repo.User,
		ExperienceRepo: repo.Experience,
		QuestionRepo:   repo.Question,
		JwtKey:         cfg.JwtSecret,
		JwtTTL:         time.Duration(cfg.JwtTTL) * time.Minute,
		OpenAI:         openaiClient,
		OpenAIModel:    cfg.OpenAIModel,
		TokenMaker:     tokenMaker,
	}

	app := &application{
		DB:         pool,
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
