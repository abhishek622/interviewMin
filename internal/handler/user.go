package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// SignUp creates a new user and returns a token
func (h *Handler) SignUp(c *gin.Context) {
	var req model.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Sugar().Warnw("signup bad request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	pwHash, err := hashPassword(req.Password)
	if err != nil {
		h.Logger.Sugar().Errorw("failed to hash password", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	user := &model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: pwHash,
	}

	createdUser, err := h.UserRepo.Create(ctx, user)
	if err != nil {
		h.Logger.Sugar().Errorw("user create failed", "email", req.Email, "err", err)
		// hide DB errors from clients; assume duplicate email will be surfaced by repo
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create user"})
		return
	}

	// generate token
	token, _, err := h.TokenMaker.GenerateToken(createdUser.UserID, createdUser.Email, createdUser.IsAdmin, h.JwtTTL)
	if err != nil {
		h.Logger.Sugar().Errorw("token generation failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	expiresAt := time.Now().Add(h.JwtTTL).Unix()

	c.JSON(http.StatusCreated, gin.H{
		"user":  model.UserResponse{UserID: createdUser.UserID, Email: createdUser.Email, Name: createdUser.Name, IsAdmin: createdUser.IsAdmin},
		"token": model.TokenResponse{AccessToken: token, ExpiresAt: expiresAt},
	})
}

// Login verifies credentials and returns JWT
func (h *Handler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Sugar().Warnw("login bad request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	user, err := h.UserRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		h.Logger.Sugar().Warnw("login user not found", "email", req.Email, "err", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if err := comparePassword(user.PasswordHash, req.Password); err != nil {
		h.Logger.Sugar().Warnw("login password mismatch", "email", req.Email, "err", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, _, err := h.TokenMaker.GenerateToken(user.UserID, user.Email, user.IsAdmin, h.JwtTTL)
	if err != nil {
		h.Logger.Sugar().Errorw("token generation failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}
	expiresAt := time.Now().Add(h.JwtTTL).Unix()

	c.JSON(http.StatusOK, gin.H{
		"user":  model.UserResponse{UserID: user.UserID, Email: user.Email, Name: user.Name, IsAdmin: user.IsAdmin},
		"token": model.TokenResponse{AccessToken: token, ExpiresAt: expiresAt},
	})
}

// Me returns the current user profile
func (h *Handler) Me(c *gin.Context) {
	user := h.GetUserFromContext(c)
	fmt.Println(user)
	if user.UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, model.UserResponse{UserID: user.UserID, Email: user.Email, Name: user.Name, IsAdmin: user.IsAdmin})
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func comparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
