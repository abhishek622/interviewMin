package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

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

const version = "1.0.0"

type application struct {
	DB         *pgxpool.Pool
	Logger     *zap.Logger
	Config     *config.Config
	Repository *repository.Repository
	Handler    *handler.Handler
}

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg.Env)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = log.Sync()
	}()

	sugar := log.Sugar()

	// Connect to database with config
	pool, err := database.Connect(ctx, &cfg.DB)
	if err != nil {
		sugar.Fatalw("failed to connect to database", "error", err)
	}
	defer pool.Close()
	sugar.Info("database connection established")

	// Initialize dependencies
	repo := repository.NewRepository(pool)
	groqClient := groq.NewClient(cfg.Groq.APIKey, cfg.Groq.Model, cfg.Groq.Timeout, log)
	tokenMaker := auth.NewJWTMaker(cfg.JWT.Secret)

	cryptoSvc, err := pkg.NewCrypto(cfg.Crypto.Secret)
	if err != nil {
		sugar.Fatalw("failed to initialize crypto service", "error", err)
	}

	hndl := handler.NewHandler(log, repo, tokenMaker, cryptoSvc, groqClient, cfg)

	app := &application{
		DB:         pool,
		Logger:     log,
		Config:     cfg,
		Repository: repo,
		Handler:    hndl,
	}

	// Graceful shutdown setup
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := app.serve(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-shutdownChan:
		sugar.Infow("received shutdown signal", "signal", sig.String())
	case err := <-errChan:
		sugar.Fatalw("server error", "error", err)
	}

	// Graceful shutdown with timeout
	sugar.Info("initiating graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.shutdown(ctx); err != nil {
		sugar.Errorw("failed to shutdown server gracefully", "error", err)
	}

	// Close database connections
	pool.Close()
	sugar.Info("database connections closed")

	sugar.Info("shutdown complete")
}
