package postgres

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool создаёт пул соединений PostgreSQL.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("разбор DSN: %w", err)
	}

	config.MaxConns = 20
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("создание пула соединений: %w", err)
	}

	// Проверка подключения
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("проверка подключения к БД: %w", err)
	}

	zap.L().Info("подключение к PostgreSQL установлено", zap.String("dsn", dsn))
	return pool, nil
}
