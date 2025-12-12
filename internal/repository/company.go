package repository

import (
	"context"
	"fmt"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/google/uuid"
)

func (r *Repository) CompanyDetails(ctx context.Context, userID uuid.UUID, companyID uuid.UUID) (*model.CompanyDetails, error) {
	const q = `SELECT c.company_id, c.name, c.slug, COUNT(i.interview_id) AS total_interviews, ROUND(AVG(COALESCE(i.no_of_round, 0)))::int AS avg_rounds
FROM companies c
LEFT JOIN interviews i ON i.company_id = c.company_id
WHERE c.user_id = $1 AND c.company_id = $2
GROUP BY c.company_id;
`
	row := r.db.QueryRow(ctx, q, userID, companyID)
	var company model.CompanyDetails
	err := row.Scan(&company.CompanyID, &company.Name, &company.Slug, &company.TotalInterviews, &company.AvgRounds)
	if err != nil {
		return nil, fmt.Errorf("query company details: %w", err)
	}
	return &company, nil
}

func (r *Repository) GetCompanyByName(ctx context.Context, userID uuid.UUID, name string) (*model.Company, error) {
	const q = `SELECT company_id, name, slug FROM companies WHERE user_id = $1 AND name = $2`
	row := r.db.QueryRow(ctx, q, userID, name)
	var company model.Company
	err := row.Scan(&company.CompanyID, &company.Name, &company.Slug)
	if err != nil {
		return nil, fmt.Errorf("query company by name: %w", err)
	}
	return &company, nil
}

func (r *Repository) CompanyList(ctx context.Context, userID uuid.UUID, limit, offset int, sort string) ([]model.CompanyList, int, error) {
	var total int
	const qTotal = `SELECT COUNT(DISTINCT c.company_id) FROM companies c INNER JOIN interviews i ON c.company_id = i.company_id WHERE c.user_id = $1`
	if err := r.db.QueryRow(ctx, qTotal, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("query company list total: %w", err)
	}

	orderBy := "MAX(i.updated_at) DESC"
	switch sort {
	case "created_at":
		orderBy = "MAX(i.updated_at) DESC"
	case "interviews":
		orderBy = "total_interviews DESC"
	case "name":
		orderBy = "c.name ASC"
	}

	q := fmt.Sprintf(`SELECT c.company_id, c.name, c.slug, COUNT(i.interview_id) AS total_interviews 
		FROM companies c 
		INNER JOIN interviews i ON c.company_id = i.company_id 
		WHERE c.user_id = $1 
		GROUP BY c.company_id 
		ORDER BY %s 
		LIMIT $2 OFFSET $3`, orderBy)

	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query company list: %w", err)
	}
	defer rows.Close()

	var out []model.CompanyList
	for rows.Next() {
		var cl model.CompanyList
		if err := rows.Scan(&cl.CompanyID, &cl.Name, &cl.Slug, &cl.TotalInterviews); err != nil {
			return nil, 0, fmt.Errorf("scan company list: %w", err)
		}
		out = append(out, cl)
	}
	return out, total, nil
}

func (r *Repository) CreateCompany(ctx context.Context, company *model.Company) (*uuid.UUID, error) {
	const q = `INSERT INTO companies (name, slug, user_id) VALUES ($1, $2, $3) RETURNING company_id`
	row := r.db.QueryRow(ctx, q, company.Name, company.Slug, company.UserID)
	var companyID uuid.UUID
	if err := row.Scan(&companyID); err != nil {
		return nil, fmt.Errorf("create company: %w", err)
	}
	return &companyID, nil
}

func (r *Repository) UpdateCompany(ctx context.Context, companyID uuid.UUID, company *model.Company) error {
	const q = `UPDATE companies SET name = $1, slug = $2 WHERE company_id = $3`
	_, err := r.db.Exec(ctx, q, company.Name, company.Slug, companyID)
	if err != nil {
		return fmt.Errorf("update company: %w", err)
	}
	return nil
}

func (r *Repository) DeleteCompany(ctx context.Context, companyID uuid.UUID) error {
	const q = `DELETE FROM companies WHERE company_id = $1`
	_, err := r.db.Exec(ctx, q, companyID)
	if err != nil {
		return fmt.Errorf("delete company: %w", err)
	}
	return nil
}

func (r *Repository) GetCompanyBySlug(ctx context.Context, userID uuid.UUID, slug string) (*model.CompanyDetails, error) {
	const q = `SELECT c.company_id, c.name, c.slug, COUNT(i.interview_id) AS total_interviews, ROUND(AVG(COALESCE(i.no_of_round, 0)))::int AS avg_rounds
FROM companies c
LEFT JOIN interviews i ON i.company_id = c.company_id
WHERE c.user_id = $1 AND c.slug = $2
GROUP BY c.company_id;
`
	row := r.db.QueryRow(ctx, q, userID, slug)
	var company model.CompanyDetails
	err := row.Scan(&company.CompanyID, &company.Name, &company.Slug, &company.TotalInterviews, &company.AvgRounds)
	if err != nil {
		return nil, fmt.Errorf("query company details: %w", err)
	}
	return &company, nil
}
