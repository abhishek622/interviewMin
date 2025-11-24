package handler

import (
	"github.com/abhishek622/interviewMin/internal/openai"
	"github.com/abhishek622/interviewMin/internal/repository"
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
