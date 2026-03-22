package repository

import (
	"Ustasjs/yp-url-shortener/internal/model"
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

var ErrConflict = errors.New("url already exists")

func (s *PostgresStorage) Save(ctx context.Context, id string, url string, userID string) error {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO short_urls (short_id, original_url, user_id)
         VALUES ($1, $2, $3)
         ON CONFLICT (original_url) DO NOTHING`,
		id, url, userID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrConflict
	}
	return nil
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

func (s *PostgresStorage) SaveListUrls(ctx context.Context, records []model.ShortURLRecord, userID string) error {
	if len(records) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO short_urls (short_id, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (original_url) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, rec := range records {
		_, err = stmt.ExecContext(ctx, rec.ID, rec.URL, userID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *PostgresStorage) CreateUser(ctx context.Context) (string, error) {
	var userID string
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO users DEFAULT VALUES RETURNING id`,
	).Scan(&userID)
	return userID, err
}

func (s *PostgresStorage) GetUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT short_id, original_url FROM short_urls WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []model.UserURLItem
	for rows.Next() {
		var item model.UserURLItem
		if err := rows.Scan(&item.ShortURL, &item.OriginalURL); err != nil {
			return nil, err
		}
		urls = append(urls, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}
