package handler

import (
	"net/http"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) CompanyDetails(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

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

	company, err := h.Repository.CompanyDetails(c.Request.Context(), claims.UserID, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "company fetched successfully", "company": company})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete company"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "company deleted successfully"})
}
