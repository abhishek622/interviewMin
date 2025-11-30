package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/abhishek622/interviewMin/internal/auth"
	"github.com/gin-gonic/gin"
)

func (app *application) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := verifyClaimsFromAuthHeader(c, app.Handler.TokenMaker)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Check if user still exists
		_, err = app.Repository.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

func (app *application) AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := verifyClaimsFromAuthHeader(c, app.Handler.TokenMaker)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		user, err := app.Repository.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil || !user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

func verifyClaimsFromAuthHeader(c *gin.Context, tokenMaker *auth.JWTMaker) (*auth.UserClaims, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header is missing")
	}

	fields := strings.Fields(authHeader)
	if len(fields) != 2 || fields[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header")
	}

	token := fields[1]
	claims, err := tokenMaker.VerifyToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}
