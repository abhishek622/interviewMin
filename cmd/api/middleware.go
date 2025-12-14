package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/abhishek622/interviewMin/internal/auth"
	"github.com/abhishek622/interviewMin/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// rateLimiter manages per-IP rate limiting with configurable limits
type rateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	rps      rate.Limit
	burst    int
	enabled  bool
}

// visitor tracks rate limiting state for a single IP
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// newRateLimiter creates a new rate limiter with the given configuration
func newRateLimiter(rps float64, burst int, enabled bool) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rps:      rate.Limit(rps),
		burst:    burst,
		enabled:  enabled,
	}

	// Start background cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// getVisitor returns or creates a rate limiter for the given IP
func (rl *rateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rps, rl.burst)
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupVisitors removes stale visitors to prevent memory leaks
func (rl *rateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// allow checks if a request from the given IP should be allowed
func (rl *rateLimiter) allow(ip string) bool {
	if !rl.enabled {
		return true
	}
	return rl.getVisitor(ip).Allow()
}

// RateLimitMiddleware creates a rate limiting middleware using app configuration
func (app *application) RateLimitMiddleware() gin.HandlerFunc {
	limiter := newRateLimiter(
		app.Config.Limiter.RPS,
		app.Config.Limiter.Burst,
		app.Config.Limiter.Enabled,
	)

	return func(c *gin.Context) {
		ip := getClientIP(c)

		if !limiter.allow(ip) {
			app.Logger.Warn("rate limit exceeded",
				zap.String("ip", ip),
				zap.String("path", c.Request.URL.Path),
			)
			response.TooManyRequests(c, "")
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientIP extracts the real client IP, handling proxies
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (common for reverse proxies)
	if xff := c.Request.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header (nginx)
	if xri := c.Request.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

var readOnlyPOSTEndpoints = map[string]bool{
	"/api/v1/interviews/list": true,
}

// ReadOnlyMiddleware restricts write operations for demo users in production
func (app *application) ReadOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.Next()
			return
		}

		userClaims, ok := claims.(*auth.UserClaims)
		if !ok {
			c.Next()
			return
		}

		// Block restricted demo user in production
		const previewUserID = "52dad4d5-261a-4dd0-8b40-46b107f7cc89"
		if userClaims.UserID.String() == previewUserID && app.Config.IsProduction() {
			if isWriteOperation(c.Request.Method, c.Request.URL.Path) {
				response.Forbidden(c, "This is a preview user. Write operations are restricted.")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// isWriteOperation checks if the HTTP request is a write operation
func isWriteOperation(method, path string) bool {
	switch method {
	case "PUT", "PATCH", "DELETE":
		return true
	case "POST":
		// Check if this POST endpoint is actually read-only
		return !readOnlyPOSTEndpoints[path]
	default:
		return false
	}
}

// AuthMiddleware validates JWT tokens and session state
func (app *application) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := verifyClaimsFromAuthHeader(c, app.Handler.TokenMaker)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// Verify session if SessionID is present
		if claims.SessionID != "" {
			session, err := app.Repository.GetUserSession(c.Request.Context(), claims.SessionID)
			if err != nil {
				app.Logger.Warn("auth: session lookup failed",
					zap.String("session_id", claims.SessionID),
					zap.Error(err),
				)
				response.Unauthorized(c, "unauthorized access")
				c.Abort()
				return
			}

			if session.IsRevoked || session.ExpiresAt.Before(time.Now().UTC()) || session.UserID != claims.UserID {
				response.Unauthorized(c, "unauthorized access")
				c.Abort()
				return
			}
		}

		c.Set("claims", claims)
		c.Next()
	}
}

// AdminMiddleware validates admin privileges
func (app *application) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := verifyClaimsFromAuthHeader(c, app.Handler.TokenMaker)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		user, err := app.Repository.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil || !user.IsAdmin {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

// verifyClaimsFromAuthHeader extracts and validates JWT claims from Authorization header
func verifyClaimsFromAuthHeader(c *gin.Context, tokenMaker *auth.JWTMaker) (*auth.UserClaims, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header is missing")
	}

	fields := strings.Fields(authHeader)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	token := fields[1]
	claims, err := tokenMaker.VerifyToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}
