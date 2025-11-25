package main

import (
	"net/http"
	"strings"

	"github.com/abhishek622/interviewMin/internal/auth"
	"github.com/gin-gonic/gin"
)

func (app *application) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(app.Config.JwtSecret, tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Inavlid token"})
			c.Abort()
			return
		}
		userId := claims.UserID
		user, err := app.Repository.User.GetByID(c.Request.Context(), userId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}
		c.Set("user", user)
		c.Next()
	}
}
