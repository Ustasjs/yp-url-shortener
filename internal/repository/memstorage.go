package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/model"

	"github.com/google/uuid"
)

// URLRecord is the in-memory representation of a stored short URL.
type URLRecord struct {
	URL       string
	UserID    string
	IsDeleted bool
}

// MemStorage is an in-memory Storage backend that persists its contents to a
// JSON file. It is safe for concurrent use.
type MemStorage struct {
	mu       sync.RWMutex
	storage  map[string]URLRecord
	users    map[string]struct{}
	filePath string
}

// NewMemStorage returns a MemStorage that persists to filePath, loading any
// previously saved data from that file.
func NewMemStorage(filePath string) *MemStorage {
	storage := &MemStorage{
		storage:  make(map[string]URLRecord),
		users:    make(map[string]struct{}),
		filePath: filePath,
	}
	err := storage.LoadFromFile()
	if err != nil {
		logger.Log.Error(err.Error())
	}
	return storage
}

// Save stores a single short URL and persists the storage to disk.
func (s *MemStorage) Save(_ctx context.Context, id string, url string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage[id] = URLRecord{URL: url, UserID: userID, IsDeleted: false}
	err := s.SaveToFile()
	if err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	return nil
}

// Get returns the original URL for id. It returns ErrRecordNotFound if the id is
// unknown and ErrDeleted if the URL has been soft-deleted.
func (s *MemStorage) Get(_ctx context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.storage[id]
	if !ok {
		return "", ErrRecordNotFound
	}
	if record.IsDeleted {
		return "", ErrDeleted
	}
	return record.URL, nil
}

// SaveListUrls stores a batch of records and persists the storage to disk.
func (s *MemStorage) SaveListUrls(_ctx context.Context, records []model.ShortURLRecord, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, rec := range records {
		s.storage[rec.ID] = URLRecord{URL: rec.URL, UserID: userID, IsDeleted: false}
	}
	return s.SaveToFile()
}

// SaveToFile writes the current contents of the storage to its backing file as
// JSON.
func (s *MemStorage) SaveToFile() error {
	data, err := json.Marshal(s.storage)
	if err != nil {
		return err
	}
	err = os.WriteFile(s.filePath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

// LoadFromFile loads the storage contents from the backing file. A missing file
// is not an error.
func (s *MemStorage) LoadFromFile() error {
	_, err := os.Stat(s.filePath)
	if errors.Is(err, os.ErrNotExist) {
		logger.Log.Info("storage file not found")
		return nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &s.storage)
	if err != nil {
		return err
	}
	return nil
}

// CreateUser registers a new user and returns a freshly generated identifier.
func (s *MemStorage) CreateUser(_ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userID := uuid.New().String()
	s.users[userID] = struct{}{}

	return userID, nil
}

// GetUserURLs returns every short/original URL pair owned by userID.
func (s *MemStorage) GetUserURLs(_ctx context.Context, userID string) ([]model.UserURLItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var urls []model.UserURLItem
	for shortID, record := range s.storage {
		if record.UserID == userID {
			urls = append(urls, model.UserURLItem{
				ShortURL:    shortID,
				OriginalURL: record.URL,
			})
		}
	}

	return urls, nil
}

// DeleteURLsBatch marks the given items as deleted when owned by the requesting
// user and persists the change.
func (s *MemStorage) DeleteURLsBatch(_ctx context.Context, items []model.DeleteItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		if record, exists := s.storage[item.ShortID]; exists {
			if record.UserID == item.UserID {
				record.IsDeleted = true
				s.storage[item.ShortID] = record
			}
		}
	}

	return s.SaveToFile()
}
