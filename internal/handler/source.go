package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// --- Request DTOs ---
type createSourceRequest struct {
	Name string `json:"name" binding:"required"`
}

// ListSources - GET /api/v1/sources
func (app *Application) ListSources(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	sources, err := app.SourceRepo.ListAll(ctx)
	if err != nil {
		app.Logger.Sugar().Errorw("list sources failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list sources"})
		return
	}
	c.JSON(http.StatusOK, sources)
}

// CreateSource - POST /api/v1/sources (admin)
// If you need RBAC, ensure middleware is applied.
func (app *Application) CreateSource(c *gin.Context) {
	var req createSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Logger.Sugar().Warnw("create source bad request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := app.SourceRepo.Create(ctx, req.Name)
	if err != nil {
		app.Logger.Sugar().Errorw("create source failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create source"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "name": req.Name})
}
