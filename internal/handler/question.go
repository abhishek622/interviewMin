package handler

import (
	"strconv"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/abhishek622/interviewMin/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateQuestion creates a new question for an interview
func (h *Handler) CreateQuestion(c *gin.Context) {
	var req model.CreateQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	createdQuestion, err := h.Repository.CreateQuestion(c.Request.Context(), &model.Question{
		InterviewID: req.InterviewID,
		Question:    req.Question,
		Type:        req.Type,
	})
	if err != nil {
		h.Logger.Error("create_question: failed to create",
			zap.Int64("interview_id", req.InterviewID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to create question")
		return
	}

	h.Logger.Info("create_question: question created",
		zap.Int64("interview_id", req.InterviewID),
		zap.Int64("question_id", createdQuestion.QID),
	)

	response.Created(c, createdQuestion)
}

// ListQuestions returns all questions for an interview grouped by type
func (h *Handler) ListQuestions(c *gin.Context) {
	var req model.ListQuestionsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

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

	questions, err := h.Repository.ListQuestionByInterviewID(c.Request.Context(), interviewID)
	if err != nil {
		h.Logger.Error("list_questions: failed to fetch",
			zap.Int64("interview_id", interviewID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to fetch questions")
		return
	}

	// Group by type
	groupedQuestions := make(map[string][]model.QuestionRes)
	for _, question := range questions {
		if _, ok := groupedQuestions[question.Type]; !ok {
			groupedQuestions[question.Type] = []model.QuestionRes{}
		}
		groupedQuestions[question.Type] = append(groupedQuestions[question.Type], question)
	}

	response.OK(c, groupedQuestions)
}

// UpdateQuestion updates an existing question
func (h *Handler) UpdateQuestion(c *gin.Context) {
	idStr := c.Param("q_id")
	if idStr == "" {
		response.BadRequest(c, "question_id is required")
		return
	}

	questionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid question_id format")
		return
	}

	var req model.UpdateQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.Repository.UpdateQuestion(c.Request.Context(), questionID, req.Question, req.Type); err != nil {
		h.Logger.Error("update_question: failed to update",
			zap.Int64("question_id", questionID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to update question")
		return
	}

	h.Logger.Info("update_question: question updated",
		zap.Int64("question_id", questionID),
	)

	response.Message(c, "question updated successfully")
}

// DeleteQuestion deletes a question
func (h *Handler) DeleteQuestion(c *gin.Context) {
	idStr := c.Param("q_id")
	if idStr == "" {
		response.BadRequest(c, "question_id is required")
		return
	}

	questionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid question_id format")
		return
	}

	if err := h.Repository.DeleteQuestion(c.Request.Context(), questionID); err != nil {
		h.Logger.Error("delete_question: failed to delete",
			zap.Int64("question_id", questionID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to delete question")
		return
	}

	h.Logger.Info("delete_question: question deleted",
		zap.Int64("question_id", questionID),
	)

	response.Message(c, "question deleted successfully")
}
