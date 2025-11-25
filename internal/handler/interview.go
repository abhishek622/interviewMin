package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
)

// --- Request DTOs ---
type createInterviewRequest struct {
	Title    string `json:"title" binding:"required"`
	RawText  string `json:"raw_text" binding:"required"`
	SourceID string `json:"source_id" binding:"required,uuid"`
}

// GetInterview retrieves a single interview by ID
// GET /api/v1/interviews/:id
func (app *Application) GetInterview(c *gin.Context) {
	interviewID := c.Param("id")
	if interviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing interview id"})
		return
	}

	user := app.GetUserFromContext(c)
	if user.ID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	interview, err := app.InterviewRepo.GetByID(ctx, interviewID)
	if err != nil {
		app.Logger.Sugar().Errorw("get interview failed", "id", interviewID, "err", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "interview not found"})
		return
	}

	// Ensure the interview belongs to the current user
	if interview.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, interview)
}

// CreateInterview stores raw interview text
// POST /api/v1/interviews
func (app *Application) CreateInterview(c *gin.Context) {
	var req createInterviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Logger.Sugar().Warnw("create interview bad request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := app.GetUserFromContext(c)
	if user.ID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ctx := c.Request.Context()
	id, err := app.InterviewRepo.Create(ctx, user.ID, req.SourceID, req.Title, req.RawText)
	if err != nil {
		app.Logger.Sugar().Errorw("interview create failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create interview"})
		return
	}
	c.JSON(http.StatusCreated, model.CreateEntryResponse{ID: id})
}

// ListInterviews returns paginated interviews for the current user
// GET /api/v1/interviews
func (app *Application) ListInterviews(c *gin.Context) {
	var q model.ListInterviewsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		app.Logger.Sugar().Warnw("list interviews bad query", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := app.GetUserFromContext(c)
	if user.ID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	limit := q.PageSize
	offset := (q.Page - 1) * q.PageSize

	ctx := c.Request.Context()
	items, total, err := app.InterviewRepo.ListByUser(ctx, user.ID, limit, offset)
	if err != nil {
		app.Logger.Sugar().Errorw("list interviews repo error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       items,
		"total":      total,
		"page":       q.Page,
		"page_size":  q.PageSize,
		"totalPages": totalPages,
	})
}

// DeleteInterview removes an interview by ID
// DELETE /api/v1/interviews/:id
func (app *Application) DeleteInterview(c *gin.Context) {
	interviewID := c.Param("id")
	if interviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing interview id"})
		return
	}

	user := app.GetUserFromContext(c)
	if user.ID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// First check if interview exists and belongs to user
	interview, err := app.InterviewRepo.GetByID(ctx, interviewID)
	if err != nil {
		app.Logger.Sugar().Errorw("delete interview: fetch failed", "id", interviewID, "err", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "interview not found"})
		return
	}

	if interview.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// TODO: Implement delete method in InterviewRepo
	// For now, return not implemented
	c.JSON(http.StatusNotImplemented, gin.H{"error": "delete not implemented"})
}
