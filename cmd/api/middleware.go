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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Inavlid token"})
			c.Abort()
			return
		}
		userId := claims.ID
		user, err := app.Repository.User.GetByID(c.Request.Context(), userId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}
		c.Set("user", &user)
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
