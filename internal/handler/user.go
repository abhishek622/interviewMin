package handler

import (
	"github.com/abhishek622/interviewMin/pkg"
	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/abhishek622/interviewMin/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SignUp creates a new user and returns a success message
func (h *Handler) SignUp(c *gin.Context) {
	var req model.SignUpReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("signup: invalid request body",
			zap.Error(err),
		)
		response.BadRequest(c, "invalid request body")
		return
	}

	ctx := c.Request.Context()
	pwHash, err := pkg.HashPassword(req.Password)
	if err != nil {
		h.Logger.Error("signup: failed to hash password",
			zap.Error(err),
		)
		response.InternalError(c, "failed to create user")
		return
	}

	user := &model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: pwHash,
	}

	userID, err := h.Repository.CreateUser(ctx, user)
	if err != nil {
		h.Logger.Error("signup: failed to create user",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		response.BadRequest(c, "could not create user")
		return
	}

	// Create default "unknown company" for the new user
	company := &model.Company{
		Name:   "unknown company",
		Slug:   "unknown-company",
		UserID: *userID,
	}
	if _, err = h.Repository.CreateCompany(ctx, company); err != nil {
		h.Logger.Error("signup: failed to create default company",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "failed to initialize user")
		return
	}

	h.Logger.Info("signup: user created successfully",
		zap.String("user_id", userID.String()),
		zap.String("email", req.Email),
	)

	response.Created(c, gin.H{"message": "user created successfully"})
}

// Login verifies credentials and returns JWT tokens
func (h *Handler) Login(c *gin.Context) {
	var req model.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("login: invalid request body",
			zap.Error(err),
		)
		response.BadRequest(c, "invalid request body")
		return
	}

	ctx := c.Request.Context()
	user, err := h.Repository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		h.Logger.Warn("login: user not found",
			zap.String("email", req.Email),
		)
		response.Unauthorized(c, "invalid credentials")
		return
	}

	if err := pkg.ComparePassword(user.PasswordHash, req.Password); err != nil {
		h.Logger.Warn("login: password mismatch",
			zap.String("email", req.Email),
		)
		response.Unauthorized(c, "invalid credentials")
		return
	}

	// Generate refresh token first to establish the session
	refreshToken, refreshClaims, err := h.TokenMaker.GenerateToken(
		user.UserID,
		user.Email,
		user.IsAdmin,
		h.Config.JWT.RefreshTokenTTL,
		"",
	)
	if err != nil {
		h.Logger.Error("login: failed to generate refresh token",
			zap.String("user_id", user.UserID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "could not generate token")
		return
	}

	// Generate access token linked to the session
	accessToken, accessClaims, err := h.TokenMaker.GenerateToken(
		user.UserID,
		user.Email,
		user.IsAdmin,
		h.Config.JWT.AccessTokenTTL,
		refreshClaims.RegisteredClaims.ID,
	)
	if err != nil {
		h.Logger.Error("login: failed to generate access token",
			zap.String("user_id", user.UserID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "could not generate token")
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
		h.Logger.Error("login: failed to create session",
			zap.String("user_id", user.UserID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "could not create session")
		return
	}

	h.Logger.Info("login: user logged in successfully",
		zap.String("user_id", user.UserID.String()),
		zap.String("session_id", session.UserTokenID),
	)

	response.OK(c, model.LoginUserRes{
		SessionID:             session.UserTokenID,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessClaims.RegisteredClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
		User:                  model.UserRes{UserID: user.UserID, Email: user.Email, Name: user.Name, IsAdmin: user.IsAdmin},
	})
}

// Me returns the current user's information
func (h *Handler) Me(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	user, err := h.Repository.GetUserByID(c.Request.Context(), claims.UserID)
	if err != nil {
		h.Logger.Warn("me: user not found",
			zap.String("user_id", claims.UserID.String()),
			zap.Error(err),
		)
		response.Unauthorized(c, "")
		return
	}

	response.OK(c, model.UserRes{
		UserID:  user.UserID,
		Name:    user.Name,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
	})
}

// Logout revokes the current session
func (h *Handler) Logout(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	if err := h.Repository.DeleteUserSession(c.Request.Context(), claims.SessionID); err != nil {
		h.Logger.Error("logout: failed to revoke session",
			zap.String("session_id", claims.SessionID),
			zap.Error(err),
		)
		response.InternalError(c, "could not revoke session")
		return
	}

	h.Logger.Info("logout: user logged out",
		zap.String("user_id", claims.UserID.String()),
		zap.String("session_id", claims.SessionID),
	)

	response.Message(c, "user logged out successfully")
}

// RenewAccessToken generates a new access token using a valid refresh token
func (h *Handler) RenewAccessToken(c *gin.Context) {
	var req model.RenewAccessTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	refreshClaims, err := h.TokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		h.Logger.Warn("renew: invalid refresh token",
			zap.Error(err),
		)
		response.Unauthorized(c, "invalid refresh token")
		return
	}

	session, err := h.Repository.GetUserSession(c.Request.Context(), refreshClaims.RegisteredClaims.ID)
	if err != nil {
		h.Logger.Warn("renew: session not found",
			zap.String("session_id", refreshClaims.RegisteredClaims.ID),
			zap.Error(err),
		)
		response.Unauthorized(c, "session not found")
		return
	}

	if session.IsRevoked {
		h.Logger.Warn("renew: session is revoked",
			zap.String("session_id", session.UserTokenID),
		)
		response.Unauthorized(c, "session blocked")
		return
	}

	if session.UserID != refreshClaims.UserID {
		h.Logger.Warn("renew: session user mismatch",
			zap.String("session_user", session.UserID.String()),
			zap.String("token_user", refreshClaims.UserID.String()),
		)
		response.Unauthorized(c, "incorrect session user")
		return
	}

	if session.UserTokenID != refreshClaims.RegisteredClaims.ID {
		response.Unauthorized(c, "mismatched session token")
		return
	}

	accessToken, accessClaims, err := h.TokenMaker.GenerateToken(
		refreshClaims.UserID,
		refreshClaims.Email,
		refreshClaims.IsAdmin,
		h.Config.JWT.AccessTokenTTL,
		refreshClaims.RegisteredClaims.ID,
	)
	if err != nil {
		h.Logger.Error("renew: failed to generate access token",
			zap.Error(err),
		)
		response.InternalError(c, "could not generate access token")
		return
	}

	response.OK(c, model.RenewAccessTokenRes{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessClaims.RegisteredClaims.ExpiresAt.Time,
	})
}

// RevokeSession revokes a user session
func (h *Handler) RevokeSession(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	if err := h.Repository.DeleteUserSession(c.Request.Context(), claims.SessionID); err != nil {
		h.Logger.Error("revoke: failed to revoke session",
			zap.String("session_id", claims.SessionID),
			zap.Error(err),
		)
		response.InternalError(c, "could not revoke session")
		return
	}

	h.Logger.Info("revoke: session revoked",
		zap.String("user_id", claims.UserID.String()),
		zap.String("session_id", claims.SessionID),
	)

	response.Message(c, "session revoked successfully")
}

// ChangePassword changes a user's password (admin only)
func (h *Handler) ChangePassword(c *gin.Context) {
	var req model.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	ctx := c.Request.Context()
	pwHash, err := pkg.HashPassword(req.NewPassword)
	if err != nil {
		h.Logger.Error("change_password: failed to hash password",
			zap.Error(err),
		)
		response.InternalError(c, "could not change password")
		return
	}

	if err = h.Repository.UpdateUserPassword(ctx, req.UserID, pwHash); err != nil {
		h.Logger.Error("change_password: failed to update password",
			zap.String("user_id", req.UserID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "could not change password")
		return
	}

	h.Logger.Info("change_password: password changed",
		zap.String("user_id", req.UserID.String()),
	)

	response.Message(c, "password changed successfully")
}
