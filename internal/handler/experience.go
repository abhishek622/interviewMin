package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/abhishek622/interviewMin/internal/openai"
	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
)

// CreateExperience handles the creation of a new interview experience
// It fetches content from the source link (if provided), or uses raw text
// Then it calls OpenAI to extract structured data
func (h *Handler) CreateExperience(c *gin.Context) {
	var req model.CreateExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := h.GetUserFromContext(c)
	if user.UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 1. Get content to process
	var contentToProcess string

	// Logic:
	// If source is personal or other, treat SourceLink as text input.
	// If source is leetcode, gfg, reddit, treat SourceLink as URL and fetch.

	isTextSource := req.Source == model.SourcePersonal || req.Source == model.SourceOther

	if isTextSource {
		contentToProcess = req.SourceLink
	} else {
		// Fetcher logic removed as it was using app.Fetcher which is not available in Handler yet.
		contentToProcess = req.SourceLink
	}

	if contentToProcess == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty content"})
		return
	}

	// 2. Call OpenAI to extract data
	extracted, err := h.extractInfo(c.Request.Context(), contentToProcess)
	if err != nil {
		h.Logger.Sugar().Errorw("extraction failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ai processing failed"})
		return
	}

	// Construct Metadata
	metadata := map[string]interface{}{
		"title":           extracted.Title,
		"full_experience": extracted.FullExperience,
	}

	exp := &model.Experience{
		UserID:     user.UserID,
		Company:    extracted.Company,
		Position:   extracted.Position,
		Source:     req.Source,
		NoOfRound:  extracted.NoOfRound,
		SourceLink: req.SourceLink,
		Location:   extracted.Location,
		Metadata:   metadata,
	}

	// Fallback for required fields if AI missed them
	if exp.Company == "" {
		exp.Company = "Unknown"
	}
	if exp.Position == "" {
		exp.Position = "Unknown"
	}

	if err := h.ExperienceRepo.Create(c.Request.Context(), exp); err != nil {
		h.Logger.Sugar().Errorw("failed to create experience", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// 5. Save Questions
	if len(extracted.Questions) > 0 {
		qs := make([]model.Question, len(extracted.Questions))
		for i, q := range extracted.Questions {
			qs[i] = model.Question{
				ExpID:    exp.ExpID,
				Question: q.Question,
				Type:     q.Type,
			}
		}
		if err := h.QuestionRepo.CreateBatch(c.Request.Context(), qs); err != nil {
			h.Logger.Sugar().Errorw("failed to save question", "err", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"exp_id": exp.ExpID})
}

func (h *Handler) ListExperiences(c *gin.Context) {
	var q model.ListExperiencesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := h.GetUserFromContext(c)
	if user.UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := q.PageSize
	if limit <= 0 {
		limit = 20
	}
	offset := (q.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	exps, total, err := h.ExperienceRepo.ListByUser(c.Request.Context(), user.UserID, limit, offset)
	if err != nil {
		h.Logger.Sugar().Warnw("create experience bad request", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      exps,
		"total":     total,
		"page":      q.Page,
		"page_size": limit,
	})
}

func (h *Handler) GetExperience(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	exp, err := h.ExperienceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
		return
	}

	// Fetch questions
	qs, err := h.QuestionRepo.ListByExperienceID(c.Request.Context(), id)
	if err != nil {
		h.Logger.Sugar().Warnw("failed to fetch questions", "err", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"experience": exp,
		"questions":  qs,
	})
}

type ExtractedData struct {
	Title     string `json:"title"`
	Company   string `json:"company"`
	Position  string `json:"position"`
	Location  string `json:"location"`
	NoOfRound int    `json:"no_of_round"`
	Questions []struct {
		Question string `json:"question"`
		Type     string `json:"type"`
	} `json:"questions"`
	FullExperience string `json:"full_experience"`
}

func (h *Handler) extractInfo(ctx context.Context, content string) (*ExtractedData, error) {
	systemMsg := `You are an expert at extracting interview experience data. 
	Output JSON only. 
	Schema:
	{
		"title": "string",
		"company": "string",
		"position": "string",
		"location": "string",
		"no_of_round": int,
		"questions": [
			{
				"question": "string",
				"type": "dsa|system_design|behavioral|other"
			}
		],
		"full_experience": "string (summary or full text)"
	}
	If a field is missing, use empty string or 0.
	`

	userPrompt := fmt.Sprintf("Extract interview experience from this text:\n\n%s", content)
	if len(userPrompt) > 10000 {
		userPrompt = userPrompt[:10000]
	}

	chatReq := openai.ChatRequest{
		Model:       h.OpenAIModel,
		Messages:    []map[string]string{{"role": "system", "content": systemMsg}, {"role": "user", "content": userPrompt}},
		MaxTokens:   2000,
		Temperature: 0.0,
	}

	respStr, err := h.OpenAI.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	var extracted ExtractedData
	if err := json.Unmarshal([]byte(respStr), &extracted); err != nil {
		return nil, fmt.Errorf("failed to parse ai response: %w", err)
	}

	return &extracted, nil
}
