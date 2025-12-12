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
	"github.com/google/uuid"
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
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("fetch failed: %v", err.Error())})
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

	// return success response immediately
	c.JSON(http.StatusOK, gin.H{"message": "Interview processing started, it will be added soon"})

	// background process
	go func(userID uuid.UUID, content string, source model.Source, meta map[string]interface{}) {
		ctx := context.Background()

		extracted, err := h.GroqClient.ExtractInterview(ctx, content)
		if err != nil {
			h.Logger.Sugar().Errorw("extraction failed", "err", err.Error())
			return
		}

		var companyID uuid.UUID
		companyName := strings.TrimSpace(strings.ToLower(extracted.Company))

		if companyName != "" {
			// find existing company
			company, _ := h.Repository.GetCompanyByName(ctx, userID, companyName)
			if company != nil {
				companyID = company.CompanyID
			}

			if companyID == uuid.Nil {
				companySlug := pkg.GenerateSlug(companyName)
				newCompany := &model.Company{
					Name:   companyName,
					Slug:   companySlug,
					UserID: userID,
				}
				newCompanyID, err := h.Repository.CreateCompany(ctx, newCompany)
				if err != nil {
					h.Logger.Sugar().Errorw("failed to create new company", "err", err.Error())
				} else if newCompanyID != nil {
					companyID = *newCompanyID
				}
			}
		}

		if companyID == uuid.Nil {
			unknownCompany, _ := h.Repository.GetCompanyByName(ctx, userID, "unknown company")
			if unknownCompany != nil {
				companyID = unknownCompany.CompanyID
			} else {
				h.Logger.Sugar().Warn("unknown company not found for user")
			}
		}

		interviewID, err := h.Repository.CreateInterview(ctx, &model.Interview{
			UserID:        userID,
			Source:        source,
			RawInput:      content,
			ProcessStatus: model.ProcessStatusCompleted,
			Metadata:      meta,
			CompanyID:     companyID,
			Position:      &extracted.Position,
			NoOfRound:     &extracted.NoOfRound,
			Location:      &extracted.Location,
		})
		if err != nil {
			h.Logger.Sugar().Errorw("failed to create interview in background", "err", err.Error())
			return
		}

		// 4. Save Questions
		if len(extracted.Questions) > 0 {
			qs := make([]model.Question, len(extracted.Questions))
			for i, q := range extracted.Questions {
				qs[i] = model.Question{
					InterviewID: *interviewID,
					Question:    q.Question,
					Type:        q.Type,
				}
			}
			if err := h.Repository.CreateQuestions(ctx, qs); err != nil {
				h.Logger.Sugar().Errorw("failed to save questions", "interview_id", *interviewID, "err", err.Error())
			}
		}

	}(claims.UserID, contentToProcess, req.Source, metadata)
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

	if req.CompanyID == nil {
		companyName := strings.ToLower(req.Company)
		companySlug := pkg.GenerateSlug(companyName)
		company, _ := h.Repository.GetCompanyByName(c.Request.Context(), claims.UserID, companyName)
		if company != nil {
			req.CompanyID = &company.CompanyID
		} else {
			newCompany := &model.Company{
				Name:   companyName,
				Slug:   companySlug,
				UserID: claims.UserID,
			}
			newCompanyID, err := h.Repository.CreateCompany(c.Request.Context(), newCompany)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			req.CompanyID = newCompanyID
		}
	}

	createObj := model.Interview{
		CompanyID:     *req.CompanyID,
		UserID:        claims.UserID,
		Source:        req.Source,
		RawInput:      req.RawInput,
		ProcessStatus: model.ProcessStatusCompleted,
		Metadata:      metadata,
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

	// background question extraction
	go func(interviewID int64, content string) {
		ctx := context.Background()
		extracted, err := h.GroqClient.InterviewQuestions(ctx, content)
		if err != nil {
			h.Logger.Sugar().Warnw("background question extraction failed", "interview_id", interviewID, "err", err.Error())
			return
		}

		if len(*extracted) > 0 {
			qs := make([]model.Question, len(*extracted))
			for i, q := range *extracted {
				qs[i] = model.Question{
					InterviewID: interviewID,
					Question:    q.Question,
					Type:        q.Type,
				}
			}
			if err := h.Repository.CreateQuestions(ctx, qs); err != nil {
				h.Logger.Sugar().Errorw("failed to save questions in background", "interview_id", interviewID, "err", err.Error())
			}
		}
	}(*interviewID, req.RawInput)
}

func (h *Handler) ListInterviews(c *gin.Context) {
	var q model.ListInterviewQuery
	if err := c.ShouldBindJSON(&q); err != nil {
		if strings.Contains(err.Error(), "CompanyID") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
			return
		}
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
		}
	}

	data, total, err := h.Repository.ListInterviewByCompany(c.Request.Context(), q.CompanyID, limit, offset, filters, q.Search)
	if err != nil {
		h.Logger.Sugar().Warnw("list interviews bad request", "err", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      data,
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
		h.Logger.Sugar().Warnw("list interviews stats bad request", "err", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, stats)
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

	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	interview, err := h.Repository.GetInterviewByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "interview not found"})
			return
		}
		h.Logger.Sugar().Errorw("failed to get interview", "id", id, "err", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	company, err := h.Repository.CompanyDetails(c.Request.Context(), claims.UserID, interview.CompanyID)
	if err != nil {
		h.Logger.Sugar().Errorw("failed to get company", "id", interview.CompanyID, "err", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	interview.CompanyName = &company.Name

	// Fetch questions
	qs, err := h.Repository.ListQuestionByInterviewID(c.Request.Context(), id)
	if err != nil {
		h.Logger.Sugar().Warnw("failed to fetch questions", "err", err.Error())
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

	// update company info
	if req.CompanyID != nil && req.Company != nil {
		company := &model.Company{
			Name: *req.Company,
			Slug: pkg.GenerateSlug(*req.Company),
		}
		if err := h.Repository.UpdateCompany(c.Request.Context(), *req.CompanyID, company); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if req.CompanyID == nil && req.Company != nil {
		company := &model.Company{
			Name:   *req.Company,
			Slug:   pkg.GenerateSlug(*req.Company),
			UserID: currInterview.UserID,
		}
		companyID, err := h.Repository.CreateCompany(c.Request.Context(), company)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		req.CompanyID = companyID
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
	if req.Position != nil {
		updates["position"] = req.Position
	}
	if req.NoOfRound != nil {
		updates["no_of_round"] = req.NoOfRound
	}
	if req.Location != nil {
		updates["location"] = req.Location
	}
	updates["company_id"] = req.CompanyID

	if err := h.Repository.UpdateInterview(c.Request.Context(), interviewID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "interview updated successfully"})
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
		h.Logger.Sugar().Errorw("failed to get interview stats", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "interview stats fetched successfully", "stats": stats})
}

func (h *Handler) RecentInterviews(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	interviews, err := h.Repository.RecentInterviews(c.Request.Context(), claims.UserID)
	if err != nil {
		h.Logger.Sugar().Errorw("failed to get recent interviews", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if len(interviews) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "interviews not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "interviews fetched successfully", "interviews": interviews})
}
