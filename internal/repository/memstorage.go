package repository

import (
	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/model"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/google/uuid"
)

type URLRecord struct {
	URL    string
	UserID string
}

type MemStorage struct {
	mu       sync.RWMutex
	storage  map[string]URLRecord
	users    map[string]struct{}
	filePath string
}

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

func (s *MemStorage) Save(_ctx context.Context, id string, url string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage[id] = URLRecord{URL: url, UserID: userID}
	err := s.SaveToFile()
	if err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	return nil
}

func (s *MemStorage) Get(_ctx context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.storage[id]
	if !ok {
		return "", ErrRecordNotFound
	}
	return record.URL, nil
}

func (s *MemStorage) SaveListUrls(_ctx context.Context, records []model.ShortURLRecord, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, rec := range records {
		s.storage[rec.ID] = URLRecord{URL: rec.URL, UserID: userID}
	}
	return s.SaveToFile()
}

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

func (s *MemStorage) CreateUser(_ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userID := uuid.New().String()
	s.users[userID] = struct{}{}

	return userID, nil
}

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
