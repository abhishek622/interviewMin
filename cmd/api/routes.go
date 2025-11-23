package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	// simple logger middleware that uses zap
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		app.Logger.Sugar().Infow("http", "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "duration", time.Since(start))
	})

	// public routes
	// v1 := r.Group("/api/v1")
	// authH := NewAuthHandler(dep) // implement
	// v1.POST("/signup", authH.SignUp)
	// v1.POST("/login", authH.Login)

	// // protected routes
	// protected := v1.Group("/")
	// protected.Use(AuthMiddleware(dep.Config.JwtSecret))
	// interviewH := NewInterviewHandler(dep)
	// protected.POST("/interviews", interviewH.CreateInterview)
	// protected.GET("/interviews", interviewH.ListInterviews)
	// protected.GET("/interviews/:id", interviewH.GetInterview)
	// protected.POST("/interviews/:id/convert", interviewH.ConvertInterview) // call openai to parse

	return r
}
