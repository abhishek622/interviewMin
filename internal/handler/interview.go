package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/abhishek622/interviewMin/internal/fetcher"
	"github.com/abhishek622/interviewMin/internal/groq"
	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateInterview(c *gin.Context) {
	var req model.CreateInterviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	inputHash, err := h.Crypto.Encrypt(req.RawInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt input"})
		return
	}

	var contentToProcess string
	var fetchedTitle string

	if req.Source == model.SourceOther || req.Source == model.SourcePersonal {
		contentToProcess = req.RawInput
	} else {
		res, err := fetcher.Fetch(req.RawInput, c.Request.UserAgent())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("fetch failed: %v", err)})
			return
		}
		contentToProcess = res.Content
		fetchedTitle = res.Title
	}

	// Construct Metadata
	metadata := map[string]interface{}{
		"title":           fetchedTitle,
		"full_experience": contentToProcess,
	}

	// save initial input in db
	expID, err := h.Repository.CreateInterview(c.Request.Context(), &model.Interview{
		UserID:        claims.UserID,
		Source:        req.Source,
		RawInput:      req.RawInput,
		InputHash:     inputHash,
		ProcessStatus: model.ProcessStatusQueued,
		Metadata:      metadata,
	})
	if err != nil {
		h.Logger.Sugar().Errorw("failed to create interview", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create interview"})
		return
	}

	// return success response
	c.JSON(http.StatusOK, gin.H{"interview_id": expID, "metadata": metadata})

	// background ai process
	go func(eID int64, content string) {
		ctx := context.Background()

		// Update status to processing
		_ = h.Repository.UpdateInterview(ctx, eID, map[string]interface{}{
			"process_status": model.ProcessStatusProcessing,
		})

		extracted, err := h.extractInfo(ctx, content)
		if err != nil {
			h.Logger.Sugar().Errorw("extraction failed", "interview_id", eID, "err", err)
			_ = h.Repository.UpdateInterview(ctx, eID, map[string]interface{}{
				"process_status": model.ProcessStatusFailed,
				"process_error":  err.Error(),
			})
			return
		}

		// Update experience with extracted data
		updates := map[string]interface{}{
			"process_status": model.ProcessStatusCompleted,
			"company":        extracted.Company,
			"position":       extracted.Position,
			"no_of_round":    extracted.NoOfRound,
			"location":       extracted.Location,
		}

		if err := h.Repository.UpdateInterview(ctx, eID, updates); err != nil {
			h.Logger.Sugar().Errorw("failed to update interview", "interview_id", eID, "err", err)
		}

		// Save Questions
		if len(extracted.Questions) > 0 {
			qs := make([]model.Question, len(extracted.Questions))
			for i, q := range extracted.Questions {
				qs[i] = model.Question{
					ExpID:    eID,
					Question: q.Question,
					Type:     q.Type,
				}
			}
			if err := h.Repository.CreateQuestions(ctx, qs); err != nil {
				h.Logger.Sugar().Errorw("failed to save questions", "exp_id", eID, "err", err)
			}
		}

	}(*expID, contentToProcess)
}

func (h *Handler) ListInterviews(c *gin.Context) {
	var q model.ListInterviewQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := q.PageSize
	if limit <= 0 {
		limit = 20
	}
	offset := max((q.Page-1)*limit, 0)

	exps, total, err := h.Repository.ListInterviewByUser(c.Request.Context(), claims.UserID, limit, offset)
	if err != nil {
		h.Logger.Sugar().Warnw("list interviews bad request", "err", err)
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

func (h *Handler) GetInterview(c *gin.Context) {
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

	exp, err := h.Repository.GetInterviewByID(c.Request.Context(), id)
	if err != nil {
		h.Logger.Sugar().Errorw("failed to get interview", "id", id, "err", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "interview not found"})
		return
	}

	// Fetch questions
	qs, err := h.Repository.ListQuestionByInterviewID(c.Request.Context(), id)
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

	chatReq := groq.ChatRequest{
		Messages:    []map[string]string{{"role": "system", "content": systemMsg}, {"role": "user", "content": userPrompt}},
		MaxTokens:   2000,
		Temperature: 0.0,
	}

	respStr, err := h.GroqClient.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	var extracted ExtractedData
	if err := json.Unmarshal([]byte(respStr), &extracted); err != nil {
		return nil, fmt.Errorf("failed to parse ai response: %w", err)
	}

	return &extracted, nil
}
