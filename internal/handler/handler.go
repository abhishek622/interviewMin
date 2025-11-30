package handler

import (
	"github.com/abhishek622/interviewMin/internal/auth"
	"github.com/abhishek622/interviewMin/internal/groq"
	"github.com/abhishek622/interviewMin/internal/repository"
	"github.com/abhishek622/interviewMin/pkg"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	Logger     *zap.Logger
	Repository *repository.Repository
	TokenMaker *auth.JWTMaker
	Crypto     *pkg.Crypto
	GroqClient *groq.Client
}

func NewHandler(logger *zap.Logger, repository *repository.Repository, tokenMaker *auth.JWTMaker, crypto *pkg.Crypto, groqClient *groq.Client) *Handler {
	return &Handler{
		Logger:     logger,
		Repository: repository,
		TokenMaker: tokenMaker,
		Crypto:     crypto,
		GroqClient: groqClient,
	}
}

// GetClaimsFromContext retrieves the current user claims from the gin context
func (h *Handler) GetClaimsFromContext(c *gin.Context) *auth.UserClaims {
	contextClaims, exists := c.Get("claims")
	if !exists {
		return nil
	}

	claims, ok := contextClaims.(*auth.UserClaims)
	if !ok {
		return nil
	}

	return claims
}
