package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/abhishek622/interviewMin/internal/fetcher"
	"github.com/abhishek622/interviewMin/pkg"
	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/abhishek622/interviewMin/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// CreateInterviewWithAI creates an interview using AI extraction from URL or text
func (h *Handler) CreateInterviewWithAI(c *gin.Context) {
	var req model.CreateInterviewWithAIReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	var contentToProcess string
	var fetchedTitle string

	if req.Source == model.SourceOther || req.Source == model.SourcePersonal {
		contentToProcess = req.RawInput
	} else {
		res, err := fetcher.Fetch(req.RawInput, req.Source, c.Request.UserAgent())
		if err != nil {
			h.Logger.Warn("create_interview_ai: fetch failed",
				zap.String("source", string(req.Source)),
				zap.Error(err),
			)
			response.BadRequest(c, "failed to fetch content from URL")
			return
		}
		contentToProcess = strings.TrimSpace(res.Content)
		fetchedTitle = strings.TrimSpace(res.Title)
	}

	// Construct Metadata
	metadata := map[string]interface{}{
		"title":           fetchedTitle,
		"full_experience": contentToProcess,
	}

	contentToProcess = fmt.Sprintf("%s\n\n%s", fetchedTitle, contentToProcess)

	unknownCompany, err := h.Repository.GetCompanyByName(c.Request.Context(), claims.UserID, "unknown company")
	if err != nil {
		h.Logger.Error("create_interview_ai: failed to get unknown company",
			zap.String("user_id", claims.UserID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "failed to get unknown company")
		return
	}
	if unknownCompany == nil {
		response.InternalError(c, "unknown company not found")
		return
	}

	response.OK(c, gin.H{"message": "Interview processing started, it will be added soon"})

	// Background process
	go func(userID uuid.UUID, content string, source model.Source, meta map[string]interface{}) {
		ctx := context.Background()

		extracted, err := h.GroqClient.ExtractInterview(ctx, content)
		if err != nil {
			h.Logger.Error("create_interview_ai: extraction failed",
				zap.String("user_id", userID.String()),
				zap.Error(err),
			)

			// create interview in unknown company
			_, err = h.Repository.CreateInterview(ctx, &model.Interview{
				UserID:        userID,
				Source:        source,
				RawInput:      content,
				ProcessStatus: model.ProcessStatusFailed,
				Metadata:      meta,
				CompanyID:     unknownCompany.CompanyID,
				ProcessError:  pkg.StringPtr(err.Error()),
			})
			if err != nil {
				h.Logger.Error("create_interview_ai: failed to create interview",
					zap.String("user_id", userID.String()),
					zap.Error(err),
				)
			}
			return
		}

		var companyID uuid.UUID
		companyName := strings.TrimSpace(strings.ToLower(extracted.Company))

		if companyName != "" {
			// Find existing company
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
					h.Logger.Error("create_interview_ai: failed to create company",
						zap.String("company_name", companyName),
						zap.Error(err),
					)
				} else if newCompanyID != nil {
					companyID = *newCompanyID
				}
			}
		}

		if companyID == uuid.Nil {
			companyID = unknownCompany.CompanyID
		}

		interviewID, err := h.Repository.CreateInterview(ctx, &model.Interview{
			UserID:        userID,
			Source:        source,
			RawInput:      content,
			ProcessStatus: model.ProcessStatusSuccess,
			Metadata:      meta,
			CompanyID:     companyID,
			Position:      &extracted.Position,
			NoOfRound:     &extracted.NoOfRound,
			Location:      &extracted.Location,
		})
		if err != nil {
			h.Logger.Error("create_interview_ai: failed to create interview",
				zap.String("user_id", userID.String()),
				zap.Error(err),
			)
			return
		}

		// Save Questions
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
				h.Logger.Error("create_interview_ai: failed to save questions",
					zap.Int64("interview_id", *interviewID),
					zap.Error(err),
				)
			}
		}

		h.Logger.Info("create_interview_ai: interview created successfully",
			zap.String("user_id", userID.String()),
			zap.Int64("interview_id", *interviewID),
		)
	}(claims.UserID, contentToProcess, req.Source, metadata)
}

// CreateInterview creates an interview with manual input
func (h *Handler) CreateInterview(c *gin.Context) {
	var req model.CreateInterviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	metadata := map[string]interface{}{
		"title":           fmt.Sprintf("%s Interview at %s", req.Position, req.Company),
		"full_experience": strings.TrimSpace(req.RawInput),
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
				h.Logger.Error("create_interview: failed to create company",
					zap.String("company_name", companyName),
					zap.Error(err),
				)
				response.InternalError(c, "failed to create company")
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
		ProcessStatus: model.ProcessStatusSuccess,
		Metadata:      metadata,
		Position:      &req.Position,
		NoOfRound:     req.NoOfRound,
		Location:      req.Location,
	}

	interviewID, err := h.Repository.CreateFullInterview(c.Request.Context(), &createObj)
	if err != nil {
		h.Logger.Error("create_interview: failed to create interview",
			zap.String("user_id", claims.UserID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "failed to create interview")
		return
	}

	h.Logger.Info("create_interview: interview created",
		zap.String("user_id", claims.UserID.String()),
		zap.Int64("interview_id", *interviewID),
	)

	response.OK(c, gin.H{
		"message":      "interview created successfully",
		"interview_id": interviewID,
	})

	// Background question extraction
	go func(interviewID int64, content string) {
		ctx := context.Background()
		extracted, err := h.GroqClient.InterviewQuestions(ctx, content)
		if err != nil {
			h.Logger.Warn("create_interview: background question extraction failed",
				zap.Int64("interview_id", interviewID),
				zap.Error(err),
			)
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
				h.Logger.Error("create_interview: failed to save questions",
					zap.Int64("interview_id", interviewID),
					zap.Error(err),
				)
			}
		}
	}(*interviewID, req.RawInput)
}

// ListInterviews returns a paginated list of interviews for a company
func (h *Handler) ListInterviews(c *gin.Context) {
	var q model.ListInterviewQuery
	if err := c.ShouldBindJSON(&q); err != nil {
		if strings.Contains(err.Error(), "CompanyID") {
			response.BadRequest(c, "company_id is required")
			return
		}
		response.BadRequest(c, "invalid request body")
		return
	}

	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	limit := q.PageSize
	if limit <= 0 {
		limit = 20
	}
	offset := max((q.Page-1)*limit, 0)

	// Build filters
	filters := make(map[string]interface{})
	if q.Filter != nil {
		if q.Filter.Source != nil {
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
		h.Logger.Error("list_interviews: failed to fetch interviews",
			zap.String("company_id", q.CompanyID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "failed to fetch interviews")
		return
	}

	response.OKWithMeta(c, data, &response.Meta{
		Page:     q.Page,
		PageSize: limit,
		Total:    total,
	})
}

// ListInterviewStats returns interview statistics for a company
func (h *Handler) ListInterviewStats(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		response.BadRequest(c, "company_id is required")
		return
	}

	uid, err := uuid.Parse(companyID)
	if err != nil {
		response.BadRequest(c, "invalid company_id format")
		return
	}

	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	stats, err := h.Repository.ListInterviewStats(c.Request.Context(), claims.UserID, uid)
	if err != nil {
		h.Logger.Error("list_interview_stats: failed to fetch stats",
			zap.String("company_id", companyID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to fetch interview stats")
		return
	}

	response.OK(c, stats)
}

// GetInterview returns a single interview with its questions
func (h *Handler) GetInterview(c *gin.Context) {
	idStr := c.Param("interview_id")
	if idStr == "" {
		response.BadRequest(c, "interview_id is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid interview_id format")
		return
	}

	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	interview, err := h.Repository.GetInterviewByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(c, "interview not found")
			return
		}
		h.Logger.Error("get_interview: failed to fetch interview",
			zap.Int64("interview_id", id),
			zap.Error(err),
		)
		response.InternalError(c, "failed to fetch interview")
		return
	}

	company, err := h.Repository.CompanyDetails(c.Request.Context(), claims.UserID, interview.CompanyID)
	if err != nil {
		h.Logger.Error("get_interview: failed to fetch company",
			zap.String("company_id", interview.CompanyID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "failed to fetch interview details")
		return
	}

	interview.CompanyName = &company.Name

	// // Fetch questions
	// questions, err := h.Repository.ListQuestionByInterviewID(c.Request.Context(), id)
	// if err != nil {
	// 	h.Logger.Warn("get_interview: failed to fetch questions",
	// 		zap.Int64("interview_id", id),
	// 		zap.Error(err),
	// 	)
	// 	questions = nil
	// }

	response.OK(c, interview)
}

// PatchInterview updates an existing interview
func (h *Handler) PatchInterview(c *gin.Context) {
	idStr := c.Param("interview_id")
	if idStr == "" {
		response.BadRequest(c, "interview_id is required")
		return
	}

	interviewID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid interview_id format")
		return
	}

	currInterview, err := h.Repository.GetInterviewByID(c.Request.Context(), interviewID)
	if err != nil {
		response.NotFound(c, "interview not found")
		return
	}

	var req model.PatchInterviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	// Update company info
	if req.CompanyID != nil && req.Company != nil {
		company := &model.Company{
			Name: *req.Company,
			Slug: pkg.GenerateSlug(*req.Company),
		}
		if err := h.Repository.UpdateCompany(c.Request.Context(), *req.CompanyID, company); err != nil {
			h.Logger.Error("patch_interview: failed to update company",
				zap.String("company_id", req.CompanyID.String()),
				zap.Error(err),
			)
			response.InternalError(c, "failed to update company")
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
			h.Logger.Error("patch_interview: failed to create company",
				zap.String("company_name", *req.Company),
				zap.Error(err),
			)
			response.InternalError(c, "failed to create company")
			return
		}
		req.CompanyID = companyID
	}

	metadata := currInterview.Metadata
	if req.Title != nil {
		metadata["title"] = req.Title
	}
	if req.RawInput != nil {
		metadata["full_experience"] = strings.TrimSpace(*req.RawInput)
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
		h.Logger.Error("patch_interview: failed to update interview",
			zap.Int64("interview_id", interviewID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to update interview")
		return
	}

	h.Logger.Info("patch_interview: interview updated",
		zap.Int64("interview_id", interviewID),
	)

	response.Message(c, "interview updated successfully")
}

// DeleteInterviews deletes multiple interviews
func (h *Handler) DeleteInterviews(c *gin.Context) {
	var req model.DeleteInterviewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	count, err := h.Repository.CheckInterviewExists(c.Request.Context(), req.InterviewIDs)
	if err != nil {
		response.NotFound(c, "interview not found")
		return
	}

	if count != len(req.InterviewIDs) {
		response.BadRequest(c, "some interview IDs are invalid")
		return
	}

	if err := h.Repository.DeleteInterviews(c.Request.Context(), req.InterviewIDs); err != nil {
		h.Logger.Error("delete_interviews: failed to delete",
			zap.Int("count", len(req.InterviewIDs)),
			zap.Error(err),
		)
		response.InternalError(c, "failed to delete interviews")
		return
	}

	h.Logger.Info("delete_interviews: interviews deleted",
		zap.Int("count", len(req.InterviewIDs)),
	)

	response.Message(c, "interviews deleted successfully")
}

// GetInterviewStats returns global interview statistics for a user
func (h *Handler) GetInterviewStats(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	stats, err := h.Repository.GetInterviewStats(c.Request.Context(), claims.UserID)
	if err != nil {
		h.Logger.Error("get_interview_stats: failed to fetch stats",
			zap.String("user_id", claims.UserID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "failed to fetch interview stats")
		return
	}

	response.OK(c, stats)
}

// RecentInterviews returns the most recent interviews for a user
func (h *Handler) RecentInterviews(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	interviews, err := h.Repository.RecentInterviews(c.Request.Context(), claims.UserID)
	if err != nil {
		h.Logger.Error("recent_interviews: failed to fetch",
			zap.String("user_id", claims.UserID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "failed to fetch recent interviews")
		return
	}

	if len(interviews) == 0 {
		response.OK(c, []interface{}{})
		return
	}

	response.OK(c, interviews)
}
