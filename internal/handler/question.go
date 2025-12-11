package handler

import (
	"net/http"
	"strconv"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateQuestion(c *gin.Context) {
	var req model.CreateQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdQuestion, err := h.Repository.CreateQuestion(c.Request.Context(), &model.Question{
		InterviewID: req.InterviewID,
		Question:    req.Question,
		Type:        req.Type,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "question created successfully", "question": createdQuestion})
}

func (h *Handler) ListQuestions(c *gin.Context) {
	var req model.ListQuestionsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	questions, err := h.Repository.ListQuestionByInterviewID(c.Request.Context(), interviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// formatted response, group by type
	groupedQuestions := make(map[string][]model.Question)
	for _, question := range questions {
		if _, ok := groupedQuestions[question.Type]; !ok {
			groupedQuestions[question.Type] = []model.Question{}
		}
		groupedQuestions[question.Type] = append(groupedQuestions[question.Type], question)
	}

	c.JSON(http.StatusOK, gin.H{"message": "questions fetched successfully", "questions": groupedQuestions})
}

func (h *Handler) UpdateQuestion(c *gin.Context) {
	idStr := c.Param("q_id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	QID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var req model.UpdateQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.Repository.UpdateQuestion(c.Request.Context(), QID, req.Question, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "question updated successfully"})
}

func (h *Handler) DeleteQuestion(c *gin.Context) {
	idStr := c.Param("q_id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	QID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	err = h.Repository.DeleteQuestion(c.Request.Context(), QID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "question deleted successfully"})
}
