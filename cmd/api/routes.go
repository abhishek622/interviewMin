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

	v1 := r.Group("/api/v1")
	{
		v1.POST("/signup", app.Handler.SignUp)
		v1.POST("/login", app.Handler.Login)
	}

	protected := v1.Group("/")
	protected.Use(app.AuthMiddleware())
	{
		protected.GET("/me", app.Handler.Me)
		protected.POST("/logout", app.Handler.Logout)

		// experience routes
		protected.POST("/experiences", app.Handler.CreateExperience)
		protected.GET("/experiences/:id", app.Handler.GetExperience)
		protected.GET("/experiences", app.Handler.ListExperiences)
	}

	session := v1.Group("/tokens")
	session.Use(app.AuthMiddleware())
	{
		session.POST("/revoke", app.Handler.RevokeSession)
		session.POST("/renew", app.Handler.RenewAccessToken)
	}

	return r
}
