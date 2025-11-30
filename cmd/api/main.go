package main

import (
	"context"

	"github.com/abhishek622/interviewMin/internal/auth"
	"github.com/abhishek622/interviewMin/internal/config"
	"github.com/abhishek622/interviewMin/internal/database"
	"github.com/abhishek622/interviewMin/internal/groq"
	"github.com/abhishek622/interviewMin/internal/handler"
	"github.com/abhishek622/interviewMin/internal/logger"
	"github.com/abhishek622/interviewMin/internal/repository"
	"github.com/abhishek622/interviewMin/pkg"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
)

type application struct {
	DB         *pgxpool.Pool
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

	repo := repository.NewRepository(pool)
	groqClient := groq.NewClient(cfg.GroqAPIKey, cfg.AIModel)
	tokenMaker := auth.NewJWTMaker(cfg.JwtSecret)
	cryptoSvc, err := pkg.NewCrypto(cfg.AesSecretKey)
	if err != nil {
		sugar.Fatal("invalid crypto key: ", zap.Error(err))
	}

	hndl := handler.NewHandler(log, repo, tokenMaker, cryptoSvc, groqClient)

	app := &application{
		DB:         pool,
		Logger:     log,
		Config:     cfg,
		Repository: repo,
		Handler:    hndl,
	}

	if err := app.serve(); err != nil {
		sugar.Fatal(err)
	}
}
