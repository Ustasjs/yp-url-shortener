package repository

import (
	"context"
	"database/sql"
	"errors"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) Save(ctx context.Context, id string, url string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO short_urls (short_id, original_url)
         VALUES ($1, $2)
         ON CONFLICT (original_url) DO NOTHING`,
		id, url,
	)
	return err
}

func (s *PostgresStorage) Get(ctx context.Context, id string) (string, error) {
	var originalURL string
	err := s.db.QueryRowContext(ctx,
		`SELECT original_url FROM short_urls WHERE short_id = $1`, id,
	).Scan(&originalURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrRecordNotFound
	}
	return originalURL, err
}
