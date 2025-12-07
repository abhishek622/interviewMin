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

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	v1 := r.Group("/api/v1")
	{
		v1.POST("/signup", app.Handler.SignUp)
		v1.POST("/login", app.Handler.Login)
		v1.POST("/tokens/renew", app.Handler.RenewAccessToken)
	}

	protected := v1.Group("/")
	protected.Use(app.AuthMiddleware())
	{
		protected.GET("/me", app.Handler.Me)
		protected.POST("/logout", app.Handler.Logout)
		protected.POST("/tokens/revoke", app.Handler.RevokeSession)

		// interview routes
		protected.POST("/interviews", app.Handler.CreateInterview)
		protected.POST("/interviews/ai", app.Handler.CreateInterviewWithAI)
		protected.GET("/interviews/stats", app.Handler.GetInterviewStats)
		protected.GET("/interviews/:id", app.Handler.GetInterview)
		protected.POST("/interviews/list", app.Handler.ListInterviews)
		protected.PATCH("/interviews/:id", app.Handler.PatchInterview)
		protected.DELETE("/interviews/:id", app.Handler.DeleteInterview)
		protected.DELETE("/interviews", app.Handler.DeleteInterviews)

		// question routes
		protected.POST("/questions", app.Handler.CreateQuestion)
		protected.GET("/questions/:interview_id", app.Handler.ListQuestions)
		protected.PUT("/questions/:q_id", app.Handler.UpdateQuestion)
		protected.DELETE("/questions/:q_id", app.Handler.DeleteQuestion)
	}

	return r
}
