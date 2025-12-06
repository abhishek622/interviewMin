package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"is_admin"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

func NewUserClaims(user_id uuid.UUID, email string, isAdmin bool, duration time.Duration, sessionID string) (*UserClaims, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("error generating token iD: %w", err)
	}

	finalSessionID := sessionID
	if finalSessionID == "" {
		finalSessionID = tokenID.String()
	}

	return &UserClaims{
		Email:     email,
		UserID:    user_id,
		IsAdmin:   isAdmin,
		SessionID: finalSessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}, nil
}
