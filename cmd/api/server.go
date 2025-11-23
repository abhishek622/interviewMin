package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (app *application) serve() error {
	port, err := strconv.Atoi(app.Config.Port)
	if err != nil {
		return fmt.Errorf("invalid port: %v", err)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Starting server on port: %d", port)

	return server.ListenAndServe()
}
