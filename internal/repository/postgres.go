package repository

import (
	"context"
	"database/sql"
	"errors"

	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/model"

	"go.uber.org/zap"
)

// PostgresStorage is a Storage backend backed by PostgreSQL.
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage returns a PostgresStorage that uses the given database
// handle.
func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

// ErrConflict is returned when a URL being saved already exists in storage.
var ErrConflict = errors.New("url already exists")

// Save inserts a single short URL. It returns ErrConflict if the original URL is
// already stored.
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

// Get returns the original URL for id. It returns ErrRecordNotFound if the id is
// unknown and ErrDeleted if the URL has been soft-deleted.
func (s *PostgresStorage) Get(ctx context.Context, id string) (string, error) {
	var originalURL string
	var isDeleted bool
	err := s.db.QueryRowContext(ctx,
		`SELECT original_url, is_deleted FROM short_urls WHERE short_id = $1`, id,
	).Scan(&originalURL, &isDeleted)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrRecordNotFound
	}
	if err != nil {
		return "", err
	}
	if isDeleted {
		return "", ErrDeleted
	}
	return originalURL, nil
}

// SaveListUrls inserts a batch of records in a single transaction, ignoring URLs
// that already exist.
func (s *PostgresStorage) SaveListUrls(ctx context.Context, records []model.ShortURLRecord, userID string) error {
	if len(records) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			logger.Log.Error("rollback transaction failed", zap.Error(rbErr))
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO short_urls (short_id, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (original_url) DO NOTHING`)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			logger.Log.Error("close prepared statement failed", zap.Error(closeErr))
		}
	}()

	for _, rec := range records {
		_, err = stmt.ExecContext(ctx, rec.ID, rec.URL, userID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// CreateUser inserts a new user row and returns its generated identifier.
func (s *PostgresStorage) CreateUser(ctx context.Context) (string, error) {
	var userID string
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO users DEFAULT VALUES RETURNING id`,
	).Scan(&userID)
	return userID, err
}

// GetUserURLs returns every short/original URL pair owned by userID.
func (s *PostgresStorage) GetUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT short_id, original_url FROM short_urls WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Log.Error("close rows failed", zap.Error(closeErr))
		}
	}()

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

// DeleteURLsBatch soft-deletes the given items, marking each matching row as
// deleted within a single transaction.
func (s *PostgresStorage) DeleteURLsBatch(ctx context.Context, items []model.DeleteItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			logger.Log.Error("rollback transaction failed", zap.Error(rbErr))
		}
	}()

	stmt, err := tx.PrepareContext(ctx,
		`UPDATE short_urls SET is_deleted = true 
         WHERE short_id = $1 AND user_id = $2`)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			logger.Log.Error("close prepared statement failed", zap.Error(closeErr))
		}
	}()

	for _, item := range items {
		_, err = stmt.ExecContext(ctx, item.ShortID, item.UserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
