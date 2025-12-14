package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope wraps all API responses in a consistent structure
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo contains error details for failed responses
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta contains pagination and other metadata
type Meta struct {
	Page     int  `json:"page,omitempty"`
	PageSize int  `json:"page_size,omitempty"`
	Total    int  `json:"total,omitempty"`
	HasNext  bool `json:"has_next,omitempty"`
}

// OK sends a successful response with data
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{
		Success: true,
		Data:    data,
	})
}

// OKWithMeta sends a successful response with data and pagination metadata
func OKWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Envelope{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a 201 response for successfully created resources
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{
		Success: true,
		Data:    data,
	})
}

// NoContent sends a 204 response with no body
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Message sends a success response with just a message
func Message(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Envelope{
		Success: true,
		Data:    gin.H{"message": message},
	})
}

// --- Error Responses ---

// errorResponse is a helper to send error responses
func errorResponse(c *gin.Context, status int, code, message string) {
	c.JSON(status, Envelope{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// BadRequest sends a 400 response
func BadRequest(c *gin.Context, message string) {
	errorResponse(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized sends a 401 response
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "unauthorized"
	}
	errorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 response
func Forbidden(c *gin.Context, message string) {
	errorResponse(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 response
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "resource not found"
	}
	errorResponse(c, http.StatusNotFound, "NOT_FOUND", message)
}

// InternalError sends a 500 response
// Note: Never expose internal error details to clients
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "internal server error"
	}
	errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// Conflict sends a 409 response
func Conflict(c *gin.Context, message string) {
	errorResponse(c, http.StatusConflict, "CONFLICT", message)
}

// TooManyRequests sends a 429 response
func TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = "rate limit exceeded, please try again later"
	}
	errorResponse(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", message)
}

// ValidationError sends a 422 response for validation failures
func ValidationError(c *gin.Context, message string) {
	errorResponse(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", message)
}
