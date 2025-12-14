package handler

import (
	"strings"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/abhishek622/interviewMin/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ListCompanies returns a paginated list of companies for the current user
func (h *Handler) ListCompanies(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	q := model.CompanyListReq{}
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	companies, total, err := h.Repository.CompanyList(c.Request.Context(), claims.UserID, q.Limit, q.Offset, q.Sort)
	if err != nil {
		h.Logger.Error("list_companies: failed to fetch companies",
			zap.String("user_id", claims.UserID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "failed to fetch companies")
		return
	}

	if companies == nil {
		companies = []model.CompanyList{}
	}

	hasNext := total > q.Limit*(q.Offset+1)

	response.OKWithMeta(c, companies, &response.Meta{
		Total:   total,
		HasNext: hasNext,
	})
}

// DeleteCompany deletes a company by ID
func (h *Handler) DeleteCompany(c *gin.Context) {
	companyID := c.Param("company_id")
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

	company, err := h.Repository.CompanyDetails(c.Request.Context(), claims.UserID, uid)
	if err != nil {
		response.NotFound(c, "company not found")
		return
	}

	if company.Name == "unknown company" {
		response.Forbidden(c, "cannot delete the default 'unknown company'")
		return
	}

	if err := h.Repository.DeleteCompany(c.Request.Context(), uid); err != nil {
		h.Logger.Error("delete_company: failed to delete",
			zap.String("company_id", companyID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to delete company")
		return
	}

	h.Logger.Info("delete_company: company deleted",
		zap.String("company_id", companyID),
	)

	response.Message(c, "company deleted successfully")
}

// GetCompany returns a company by ID or slug
func (h *Handler) GetCompany(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	identifier := c.Param("identifier")
	if identifier == "" {
		response.BadRequest(c, "identifier is required")
		return
	}

	// Check if identifier is UUID
	if id, err := uuid.Parse(identifier); err == nil {
		company, err := h.Repository.CompanyDetails(c.Request.Context(), claims.UserID, id)
		if err != nil {
			h.Logger.Error("get_company: failed to fetch by ID",
				zap.String("company_id", identifier),
				zap.Error(err),
			)
			response.NotFound(c, "company not found")
			return
		}
		response.OK(c, company)
		return
	}

	// Fetch by slug
	company, err := h.Repository.GetCompanyBySlug(c.Request.Context(), claims.UserID, identifier)
	if err != nil {
		h.Logger.Error("get_company: failed to fetch by slug",
			zap.String("slug", identifier),
			zap.Error(err),
		)
		response.NotFound(c, "company not found")
		return
	}

	response.OK(c, company)
}

// ListCompaniesNameList returns a list of company names for autocomplete
func (h *Handler) ListCompaniesNameList(c *gin.Context) {
	claims := h.GetClaimsFromContext(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	q := model.CompanyNameListReq{}
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	if q.Search != nil && *q.Search != "" {
		*q.Search = strings.ToLower(strings.TrimSpace(*q.Search))
	}

	companies, total, err := h.Repository.CompanyListNameList(c.Request.Context(), claims.UserID, q.Limit, q.Offset, q.Search)
	if err != nil {
		h.Logger.Error("list_companies_name_list: failed to fetch",
			zap.String("user_id", claims.UserID.String()),
			zap.Error(err),
		)
		response.InternalError(c, "failed to fetch companies")
		return
	}

	hasNext := total > q.Limit*(q.Offset+1)

	response.OKWithMeta(c, companies, &response.Meta{
		Total:   total,
		HasNext: hasNext,
	})
}
