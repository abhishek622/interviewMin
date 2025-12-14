package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// server holds the HTTP server instance for graceful shutdown
var server *http.Server

func (app *application) serve() error {
	addr := app.Config.GetServerAddr()

	server = &http.Server{
		Addr:         addr,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     zap.NewStdLog(app.Logger),
	}

	app.Logger.Info("starting server",
		zap.String("addr", addr),
		zap.String("env", app.Config.Env),
	)

	return server.ListenAndServe()
}

// shutdown gracefully shuts down the server
func (app *application) shutdown(ctx context.Context) error {
	if server == nil {
		return nil
	}

	app.Logger.Info("shutting down server...")

	if err := server.Shutdown(ctx); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("server shutdown error: %w", err)
	}

	app.Logger.Info("server stopped")
	return nil
}
