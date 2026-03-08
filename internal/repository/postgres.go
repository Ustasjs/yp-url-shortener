package repository

import (
	"Ustasjs/yp-url-shortener/internal/logger"
	"context"
	"database/sql"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) Save(ctx context.Context, id string, url string) error {
	logger.Log.Info("save to postgres")
	return nil
}

func (s *PostgresStorage) Get(ctx context.Context, id string) (string, error) {
	return "ololo", nil
}
