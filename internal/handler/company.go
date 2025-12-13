package handler

import (
	"net/http"
	"strings"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

	companies, total, err := h.Repository.CompanyList(c.Request.Context(), claims.UserID, q.Limit, q.Offset, q.Sort)
	if err != nil {
		h.Logger.Sugar().Errorw("failed to fetch company list", "err", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch company list"})
		return
	}

	hasNext := total > q.Limit*(q.Offset+1)

	c.JSON(http.StatusOK, gin.H{"message": "company list fetched successfully", "companies": companies, "has_next": hasNext})
}

func (h *Handler) DeleteCompany(c *gin.Context) {
	companyId := c.Param("company_id")
	if companyId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing company ID"})
		return
	}

	uid, err := uuid.Parse(companyId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company ID"})
		return
	}

	if err := h.Repository.DeleteCompany(c.Request.Context(), uid); err != nil {
		h.Logger.Sugar().Errorw("failed to delete company", "company_id", companyId, "err", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete company"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "company deleted successfully"})
}

func (h *Handler) GetCompany(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	identifier := c.Param("identifier")
	if identifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing identifier"})
		return
	}

	// Check if identifier is UUID
	if id, err := uuid.Parse(identifier); err == nil {
		// It is a UUID, fetch details by ID
		company, err := h.Repository.CompanyDetails(c.Request.Context(), claims.UserID, id)
		if err != nil {
			h.Logger.Sugar().Errorw("failed to fetch company details", "company_id", identifier, "err", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch company details"})
			return
		}
		c.JSON(http.StatusOK, company)
		return
	}

	// It is a slug (or invalid UUID), fetch by Slug
	company, err := h.Repository.GetCompanyBySlug(c.Request.Context(), claims.UserID, identifier)
	if err != nil {
		h.Logger.Sugar().Errorw("failed to fetch company details", "company_id", identifier, "err", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch company details"})
		return
	}
	c.JSON(http.StatusOK, company)
}

func (h *Handler) ListCompaniesNameList(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	q := model.CompanyNameListReq{}
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if q.Search != nil && *q.Search != "" {
		*q.Search = strings.ToLower(strings.TrimSpace(*q.Search))
	}

	companies, total, err := h.Repository.CompanyListNameList(c.Request.Context(), claims.UserID, q.Limit, q.Offset, q.Search)
	if err != nil {
		h.Logger.Sugar().Errorw("failed to fetch company list", "err", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch company list"})
		return
	}

	hasNext := total > q.Limit*(q.Offset+1)

	c.JSON(http.StatusOK, gin.H{"message": "company list fetched successfully", "companies": companies, "has_next": hasNext})
}
