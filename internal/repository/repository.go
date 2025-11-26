package repository

import "github.com/jackc/pgx/v5/pgxpool"

type Repository struct {
	User       UserRepository
	Experience ExperienceRepository
	Question   QuestionRepository
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		User:       UserRepository{db: db},
		Experience: ExperienceRepository{db: db},
		Question:   QuestionRepository{db: db},
	}
}
