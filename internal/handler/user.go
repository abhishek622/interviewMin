package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/abhishek622/interviewMin/pkg"
	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
)

// SignUp creates a new user and returns a token
func (h *Handler) SignUp(c *gin.Context) {
	var req model.SignUpReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Sugar().Warnw("signup bad request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	pwHash, err := pkg.HashPassword(req.Password)
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

	err = h.Repository.CreateUser(ctx, user)
	if err != nil {
		h.Logger.Sugar().Errorw("user create failed", "email", req.Email, "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})
}

// Login verifies credentials and returns JWT
func (h *Handler) Login(c *gin.Context) {
	var req model.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Sugar().Warnw("login bad request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	user, err := h.Repository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		h.Logger.Sugar().Warnw("login user not found", "email", req.Email, "err", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if err := pkg.ComparePassword(user.PasswordHash, req.Password); err != nil {
		h.Logger.Sugar().Warnw("login password mismatch", "email", req.Email, "err", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate Refresh Token first to establish the session
	refreshToken, refreshClaims, err := h.TokenMaker.GenerateToken(user.UserID, user.Email, user.IsAdmin, 24*time.Hour, "")
	if err != nil {
		h.Logger.Sugar().Errorw("error creating token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	// Generate Access Token linked to the session
	accessToken, accessClaims, err := h.TokenMaker.GenerateToken(user.UserID, user.Email, user.IsAdmin, 60*time.Minute, refreshClaims.RegisteredClaims.ID)
	if err != nil {
		h.Logger.Sugar().Errorw("error creating token", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	session, err := h.Repository.CreateUserSession(ctx, &model.UserToken{
		UserTokenID:  refreshClaims.RegisteredClaims.ID,
		UserID:       user.UserID,
		RefreshToken: refreshToken,
		ExpiresAt:    refreshClaims.RegisteredClaims.ExpiresAt.Time,
		DeviceInfo:   c.Request.UserAgent(),
		IsRevoked:    false,
	})
	if err != nil {
		h.Logger.Sugar().Errorw("error creating session", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}

	res := model.LoginUserRes{
		SessionID:             session.UserTokenID,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessClaims.RegisteredClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
		User:                  model.UserRes{UserID: user.UserID, Email: user.Email, Name: user.Name, IsAdmin: user.IsAdmin},
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (h *Handler) Me(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.Repository.GetUserByID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, model.UserRes{UserID: user.UserID, Name: user.Name, Email: user.Email, IsAdmin: user.IsAdmin})
}

func (h *Handler) Logout(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	fmt.Println(claims.RegisteredClaims.ID)
	// Use the SessionID from the claims (which links to the refresh token ID)
	err := h.Repository.DeleteUserSession(c.Request.Context(), claims.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user logged out successfully"})
}

func (h *Handler) RenewAccessToken(c *gin.Context) {
	var req model.RenewAccessTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refreshClaims, err := h.TokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	session, err := h.Repository.GetUserSession(c.Request.Context(), refreshClaims.RegisteredClaims.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if session.IsRevoked {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session blocked"})
		return
	}

	if session.UserID != refreshClaims.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect session user"})
		return
	}

	if session.UserTokenID != refreshClaims.RegisteredClaims.ID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "mismatched session token"})
		return
	}

	if time.Now().After(session.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "expired session"})
		return
	}

	accessToken, accessClaims, err := h.TokenMaker.GenerateToken(refreshClaims.UserID, refreshClaims.Email, refreshClaims.IsAdmin, 60*time.Minute, refreshClaims.RegisteredClaims.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate access token"})
		return
	}

	c.JSON(http.StatusOK, model.RenewAccessTokenRes{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessClaims.RegisteredClaims.ExpiresAt.Time,
	})
}

func (h *Handler) RevokeSession(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := h.Repository.DeleteUserSession(c.Request.Context(), claims.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "session revoked successfully"})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	var req model.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	pwHash, err := pkg.HashPassword(req.NewPassword)
	if err != nil {
		h.Logger.Sugar().Errorw("error hashing password", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not change password"})
		return
	}

	err = h.Repository.UpdateUserPassword(ctx, req.UserID, pwHash)
	if err != nil {
		h.Logger.Sugar().Errorw("error updating password", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
