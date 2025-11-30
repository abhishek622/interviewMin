package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is the concrete implementation for users.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// execTx executes a function within a database transaction.
func (r *Repository) execTx(ctx context.Context, fn func(pgx.Tx) error) error {
	// 1. Begin the transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	// 2. Defer Rollback
	// In pgx, it is safe to call Rollback on a committed transaction (it returns ErrTxClosed),
	// so we can blindly defer it for safety against panics or early returns.
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 3. Execute the logic
	if err := fn(tx); err != nil {
		return err // The defer will handle the rollback
	}

	// 4. Commit
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}
