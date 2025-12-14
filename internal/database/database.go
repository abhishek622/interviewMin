package database

import (
	"context"
	"fmt"

	"github.com/abhishek622/interviewMin/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect establishes a connection pool to the database using the provided configuration
func Connect(ctx context.Context, dbCfg *config.DBConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(dbCfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Apply connection pool settings from config
	poolConfig.MaxConns = int32(dbCfg.MaxOpenConns)
	poolConfig.MinConns = int32(dbCfg.MaxIdleConns)
	poolConfig.MaxConnIdleTime = dbCfg.MaxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// HealthCheck performs a simple health check on the database connection
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
