package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/abhishek622/interviewMin/internal/auth"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Rate Limiter Logic
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

// Example cleanup task (run in a goroutine in main if robust cleanup is needed)
func init() {
	go cleanupVisitors()
}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(20, 5) // 20 requests per second, burst of 5
		visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

func (app *application) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			ip = c.Request.RemoteAddr // Fallback if no port
		}

		// Handle X-Forwarded-For if behind a proxy
		xfwd := c.Request.Header.Get("X-Forwarded-For")
		if xfwd != "" {
			ip = strings.Split(xfwd, ",")[0]
		}
		ip = strings.TrimSpace(ip)

		limiter := getVisitor(ip)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}

		c.Next()
	}
}

func (app *application) ReadOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			// safe check
			c.Next()
			return
		}

		userClaims, ok := claims.(*auth.UserClaims)
		if !ok {
			c.Next()
			return
		}

		// Block restricted user
		if userClaims.UserID.String() == "36b584a3-f828-4925-8fdc-8fef8ae4e95a" && app.Config.Env == "production" {
			if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" || c.Request.Method == "DELETE" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "This is a demo user. Write operations are restricted."})
				return
			}
		}

		c.Next()
	}
}

func (app *application) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := verifyClaimsFromAuthHeader(c, app.Handler.TokenMaker)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Verify session if SessionID is present (it should be for access tokens)
		if claims.SessionID != "" {
			session, err := app.Repository.GetUserSession(c.Request.Context(), claims.SessionID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
				c.Abort()
				return
			}

			if session.IsRevoked || session.ExpiresAt.Before(time.Now().UTC()) || session.UserID != claims.UserID {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
				c.Abort()
				return
			}
		}

		c.Set("claims", claims)
		c.Next()
	}
}

func (app *application) AdminMiddleware() gin.HandlerFunc {
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
