package main

import (
	"net/http"
	"time"

	"github.com/abhishek622/interviewMin/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (app *application) routes() http.Handler {
	// Set Gin mode based on environment
	if app.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// Custom logging middleware using zap
	r.Use(app.loggingMiddleware())
	r.Use(app.corsMiddleware())

	// Public routes
	r.GET("/v1/healthcheck", app.healthcheckHandler)

	// API v1 routes
	v1 := r.Group("/api/v1")
	v1.Use(app.RateLimitMiddleware())
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", app.Handler.Login)
			auth.POST("/tokens/renew", app.Handler.RenewAccessToken)
		}

		protected := v1.Group("/")
		protected.Use(app.AuthMiddleware())
		protected.Use(app.ReadOnlyMiddleware())
		{
			user := protected.Group("/user")
			{
				user.GET("/me", app.Handler.Me)
				user.POST("/logout", app.Handler.Logout)
				user.POST("/tokens/revoke", app.Handler.RevokeSession)
			}

			interviews := protected.Group("/interviews")
			{
				interviews.POST("", app.Handler.CreateInterview)
				interviews.POST("/ai", app.Handler.CreateInterviewWithAI)
				interviews.POST("/list", app.Handler.ListInterviews)
				interviews.DELETE("", app.Handler.DeleteInterviews)

				interviews.GET("/stats", app.Handler.GetInterviewStats)
				interviews.GET("/recent", app.Handler.RecentInterviews)
				interviews.GET("/list/stats", app.Handler.ListInterviewStats)

				interviews.GET("/:interview_id", app.Handler.GetInterview)
				interviews.PATCH("/:interview_id", app.Handler.PatchInterview)
			}

			companies := protected.Group("/companies")
			{
				companies.GET("", app.Handler.ListCompanies)
				companies.GET("/list/names", app.Handler.ListCompaniesNameList)
				companies.GET("/:identifier", app.Handler.GetCompany)
				companies.DELETE("/:company_id", app.Handler.DeleteCompany)
			}

			questions := protected.Group("/questions")
			{
				questions.POST("", app.Handler.CreateQuestion)
				questions.GET("/:interview_id", app.Handler.ListQuestions)
				questions.PUT("/:q_id", app.Handler.UpdateQuestion)
				questions.DELETE("/:q_id", app.Handler.DeleteQuestion)
			}
		}

		admin := v1.Group("/admin")
		admin.Use(app.AdminMiddleware())
		{
			admin.POST("/signup", app.Handler.SignUp)
			admin.POST("/change-password", app.Handler.ChangePassword)
		}
	}

	return r
}

// healthcheckHandler returns the application health status and system information
func (app *application) healthcheckHandler(c *gin.Context) {
	response.OK(c, gin.H{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.Config.Env,
			"version":     version,
		},
	})
}

// loggingMiddleware provides structured HTTP request logging using zap
func (app *application) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// Use appropriate log level based on status code
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		switch {
		case statusCode >= 500:
			app.Logger.Error("http", fields...)
		case statusCode >= 400:
			app.Logger.Warn("http", fields...)
		default:
			app.Logger.Info("http", fields...)
		}
	}
}

// corsMiddleware handles Cross-Origin Resource Sharing using config
func (app *application) corsMiddleware() gin.HandlerFunc {
	allowedOrigins := make(map[string]bool)
	for _, origin := range app.Config.GetCORSOrigins() {
		allowedOrigins[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
