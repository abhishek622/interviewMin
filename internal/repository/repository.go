package repository

import "github.com/jackc/pgx/v5/pgxpool"

type Repository struct {
	User      UserRepository
	Interview InterviewRepository
	Entry     EntryRepository
	Source    SourceRepository
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		User:      UserRepository{db: db},
		Interview: InterviewRepository{db: db},
		Entry:     EntryRepository{db: db},
		Source:    SourceRepository{db: db},
	}
}
