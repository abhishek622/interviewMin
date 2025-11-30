package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/abhishek622/interviewMin/internal/auth"
	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
)

// SignUp creates a new user and returns a token
func (app *Application) SignUp(c *gin.Context) {
	var req model.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Logger.Sugar().Warnw("signup bad request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	pwHash, err := hashPassword(req.Password)
	if err != nil {
		app.Logger.Sugar().Errorw("failed to hash password", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	user, err := app.UserRepo.Create(ctx, req.Email, pwHash)
	if err != nil {
		app.Logger.Sugar().Errorw("user create failed", "email", req.Email, "err", err)
		// hide DB errors from clients; assume duplicate email will be surfaced by repo
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create user"})
		return
	}

	// generate token
	token, err := auth.GenerateToken(app.JwtKey, user.UserID, app.JwtTTL)
	if err != nil {
		app.Logger.Sugar().Errorw("token generation failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	expiresAt := time.Now().Add(time.Duration(app.JwtTTL) * time.Minute).Unix()

	c.JSON(http.StatusCreated, gin.H{
		"user":  model.UserResponse{UserID: user.UserID, Email: req.Email, Role: user.Role},
		"token": model.TokenResponse{AccessToken: token, ExpiresAt: expiresAt},
	})
}

// Login verifies credentials and returns JWT
func (app *Application) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Logger.Sugar().Warnw("login bad request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	user, err := app.UserRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		app.Logger.Sugar().Warnw("login user not found", "email", req.Email, "err", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if err := comparePassword(user.PasswordHash, req.Password); err != nil {
		app.Logger.Sugar().Warnw("login password mismatch", "email", req.Email, "err", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(app.JwtKey, user.UserID, app.JwtTTL)
	if err != nil {
		app.Logger.Sugar().Errorw("token generation failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}
	expiresAt := time.Now().Add(time.Duration(app.JwtTTL) * time.Minute).Unix()

	c.JSON(http.StatusOK, gin.H{
		"user":  model.UserResponse{UserID: user.UserID, Email: user.Email, Role: user.Role},
		"token": model.TokenResponse{AccessToken: token, ExpiresAt: expiresAt},
	})
}

// Me returns the current user profile
func (app *Application) Me(c *gin.Context) {
	user := app.GetUserFromContext(c)
	fmt.Println(user)
	if user.UserID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, model.UserResponse{UserID: user.UserID, Email: user.Email, Role: user.Role})
}
