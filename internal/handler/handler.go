package handler

import (
	"github.com/abhishek622/interviewMin/internal/auth"
	"github.com/abhishek622/interviewMin/internal/openai"
	"github.com/abhishek622/interviewMin/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	Logger      *zap.Logger
	Repository  *repository.Repository
	OpenAI      *openai.Client
	OpenAIModel string
	TokenMaker  *auth.JWTMaker
}

func NewHandler(logger *zap.Logger, repository *repository.Repository, openai *openai.Client, openaiModel string, tokenMaker *auth.JWTMaker) *Handler {
	return &Handler{
		Logger:      logger,
		Repository:  repository,
		OpenAI:      openai,
		OpenAIModel: openaiModel,
		TokenMaker:  tokenMaker,
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
