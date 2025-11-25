package handler

import (
	"github.com/abhishek622/interviewMin/internal/openai"
	"github.com/abhishek622/interviewMin/internal/repository"
	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Application struct {
	Logger        *zap.Logger
	UserRepo      repository.UserRepository
	InterviewRepo repository.InterviewRepository
	EntryRepo     repository.EntryRepository
	SourceRepo    repository.SourceRepository
	JwtKey        string
	JwtTTL        int
	OpenAI        *openai.Client
	OpenAIModel   string
}

// GetUserFromContext retrieves the current user from the gin context
func (app *Application) GetUserFromContext(c *gin.Context) *model.User {
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
