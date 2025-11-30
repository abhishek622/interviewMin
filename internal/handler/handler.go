package handler

import (
	"time"

	"github.com/abhishek622/interviewMin/internal/auth"
	"github.com/abhishek622/interviewMin/internal/openai"
	"github.com/abhishek622/interviewMin/internal/repository"
	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	Logger         *zap.Logger
	UserRepo       repository.UserRepository
	ExperienceRepo repository.ExperienceRepository
	QuestionRepo   repository.QuestionRepository
	JwtKey         string
	JwtTTL         time.Duration
	OpenAI         *openai.Client
	OpenAIModel    string
	TokenMaker     *auth.JWTMaker
}

// GetUserFromContext retrieves the current user from the gin context
func (h *Handler) GetUserFromContext(c *gin.Context) *model.User {
	contextUser, exists := c.Get("user")
	if !exists {
		return &model.User{}
	}

	user, ok := contextUser.(*model.User)
	if !ok {
		return &model.User{}
	}

	return user
}
