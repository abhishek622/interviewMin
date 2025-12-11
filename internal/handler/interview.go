package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/abhishek622/interviewMin/internal/fetcher"
	"github.com/abhishek622/interviewMin/pkg"
	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) CreateInterviewWithAI(c *gin.Context) {
	var req model.CreateInterviewWithAIReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var contentToProcess string
	var fetchedTitle string

	if req.Source == model.SourceOther || req.Source == model.SourcePersonal {
		contentToProcess = req.RawInput
	} else {
		res, err := fetcher.Fetch(req.RawInput, req.Source, c.Request.UserAgent())
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
	// attach title to contentToProcess
	contentToProcess = fmt.Sprintf("%s\n\n%s", fetchedTitle, contentToProcess)
	// fmt.Println(metadata)
	// save initial input in db
	interviewID, err := h.Repository.CreateInterview(c.Request.Context(), &model.Interview{
		UserID:        claims.UserID,
		Source:        req.Source,
		RawInput:      req.RawInput,
		ProcessStatus: model.ProcessStatusQueued,
		Metadata:      metadata,
	})
	if err != nil {
		h.Logger.Sugar().Errorw("failed to create interview", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create interview"})
		return
	}

	// return success response
	c.JSON(http.StatusOK, gin.H{"interview_id": interviewID, "metadata": metadata})

	// background ai process
	go func(interviewID int64, content string) {
		ctx := context.Background()

		// Update status to processing
		_ = h.Repository.UpdateInterview(ctx, interviewID, map[string]interface{}{
			"process_status": model.ProcessStatusProcessing,
		})

		extracted, err := h.GroqClient.ExtractInterview(ctx, content)
		if err != nil {
			h.Logger.Sugar().Errorw("extraction failed", "interview_id", interviewID, "err", err)
			_ = h.Repository.UpdateInterview(ctx, interviewID, map[string]interface{}{
				"process_status": model.ProcessStatusFailed,
				"process_error":  err.Error(),
				"attempted":      1,
			})
			return
		}
		// fmt.Println(extracted)
		// Update experience with extracted data
		updates := map[string]interface{}{
			"process_status": model.ProcessStatusCompleted,
			"company":        strings.ToLower(extracted.Company),
			"slug":           pkg.GenerateSlug(extracted.Company),
			"position":       extracted.Position,
			"no_of_round":    extracted.NoOfRound,
			"location":       extracted.Location,
			"attempted":      1,
		}

		if err := h.Repository.UpdateInterview(ctx, interviewID, updates); err != nil {
			h.Logger.Sugar().Errorw("failed to update interview", "interview_id", interviewID, "err", err)
		}

		// Save Questions
		if len(extracted.Questions) > 0 {
			qs := make([]model.Question, len(extracted.Questions))
			for i, q := range extracted.Questions {
				qs[i] = model.Question{
					InterviewID: interviewID,
					Question:    q.Question,
					Type:        q.Type,
				}
			}
			if err := h.Repository.CreateQuestions(ctx, qs); err != nil {
				h.Logger.Sugar().Errorw("failed to save questions", "interview_id", interviewID, "err", err)
			}
		}

	}(*interviewID, contentToProcess)
}

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

	metadata := map[string]interface{}{
		"title":           fmt.Sprintf("%s Interview at %s", req.Position, req.Company),
		"full_experience": req.RawInput,
	}

	req.Company = strings.ToLower(req.Company)

	createObj := model.Interview{
		UserID:        claims.UserID,
		Source:        req.Source,
		RawInput:      req.RawInput,
		ProcessStatus: model.ProcessStatusCompleted,
		Metadata:      metadata,
		Company:       &req.Company,
		Slug:          pkg.GenerateSlug(req.Company),
		Position:      &req.Position,
		NoOfRound:     req.NoOfRound,
		Location:      req.Location,
	}

	interviewID, err := h.Repository.CreateFullInterview(c.Request.Context(), &createObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "interview created successfully", "interview_id": interviewID})
}

func (h *Handler) ListInterviews(c *gin.Context) {
	var q model.ListInterviewQuery
	if err := c.ShouldBindJSON(&q); err != nil {
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

	// filters
	filters := make(map[string]interface{})

	if q.Filter != nil {
		if q.Filter.Source != nil {
			// Create a slice of standard strings
			sourceStrings := make([]string, len(*q.Filter.Source))
			for i, v := range *q.Filter.Source {
				sourceStrings[i] = string(v)
			}
			filters["source"] = sourceStrings
		}

		if q.Filter.ProcessStatus != nil {
			statusStrings := make([]string, len(*q.Filter.ProcessStatus))
			for i, v := range *q.Filter.ProcessStatus {
				statusStrings[i] = string(v)
			}
			filters["process_status"] = statusStrings
		} else if q.Filter.Status != nil {
			statusStrings := make([]string, len(*q.Filter.Status))
			for i, v := range *q.Filter.Status {
				statusStrings[i] = string(v)
			}
			filters["process_status"] = statusStrings
		}
	}

	exps, total, err := h.Repository.ListInterviewByUser(c.Request.Context(), claims.UserID, limit, offset, filters, q.Search, q.Company)
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

func (h *Handler) ListInterviewStats(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.Repository.ListInterviewByUserStats(c.Request.Context(), claims.UserID)
	if err != nil {
		h.Logger.Sugar().Warnw("list interviews stats bad request", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
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

	interview, err := h.Repository.GetInterviewByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "interview not found"})
			return
		}
		h.Logger.Sugar().Errorw("failed to get interview", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Fetch questions
	qs, err := h.Repository.ListQuestionByInterviewID(c.Request.Context(), id)
	if err != nil {
		h.Logger.Sugar().Warnw("failed to fetch questions", "err", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"interview": interview,
		"questions": qs,
	})
}

func (h *Handler) PatchInterview(c *gin.Context) {
	idStr := c.Param("interview_id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	interviewID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	currInterview, err := h.Repository.GetInterviewByID(c.Request.Context(), interviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "interview not found"})
		return
	}

	var req model.PatchInterviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metadata := currInterview.Metadata
	if req.Title != nil {
		metadata["title"] = req.Title
	}
	if req.FullExperience != nil {
		metadata["full_experience"] = req.FullExperience
	}

	updates := make(map[string]interface{})
	if len(metadata) > 0 {
		updates["metadata"] = metadata
	}
	if req.Company != nil {
		updates["company"] = strings.ToLower(*req.Company)
		updates["slug"] = pkg.GenerateSlug(*req.Company)
	}
	if req.Position != nil {
		updates["position"] = req.Position
	}
	if req.NoOfRound != nil {
		updates["no_of_round"] = req.NoOfRound
	}
	if req.Location != nil {
		updates["location"] = req.Location
	}

	if err := h.Repository.UpdateInterview(c.Request.Context(), interviewID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "interview updated successfully"})
}

func (h *Handler) DeleteInterview(c *gin.Context) {
	idStr := c.Param("interview_id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	interviewID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	count, err := h.Repository.CheckInterviewExists(c.Request.Context(), []int64{interviewID})
	if err != nil || count != 1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "interview not found"})
		return
	}

	if err := h.Repository.DeleteInterview(c.Request.Context(), interviewID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "interview deleted successfully"})
}

func (h *Handler) DeleteInterviews(c *gin.Context) {
	var req model.DeleteInterviewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	count, err := h.Repository.CheckInterviewExists(c.Request.Context(), req.InterviewIDs)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "interview not found"})
		return
	}

	if count != len(req.InterviewIDs) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid interview IDs present in list"})
		return
	}

	if err := h.Repository.DeleteInterviews(c.Request.Context(), req.InterviewIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "interviews deleted successfully"})
}

func (h *Handler) GetInterviewStats(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.Repository.GetInterviewStats(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "interview not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "interview stats fetched successfully", "stats": stats})
}

func (h *Handler) ListCompanies(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	q := model.CompanyListReq{}
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companies, total, err := h.Repository.CompanyList(c.Request.Context(), claims.UserID, q.Limit, q.Offset)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company list not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "company list fetched successfully", "companies": companies, "total": total})
}

func (h *Handler) DeleteInterviewsByCompany(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company name"})
		return
	}

	interviews, err := h.Repository.GetInterviewsByCompany(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "interviews not found"})
		return
	}

	if len(interviews) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "interviews not found for company"})
		return
	}

	if err := h.Repository.DeleteInterviews(c.Request.Context(), interviews); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "interviews deleted successfully for company"})
}

func (h *Handler) RecentInterviews(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	interviews, err := h.Repository.RecentInterviews(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "interviews not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "interviews fetched successfully", "interviews": interviews})
}
