package repository

import (
	"Ustasjs/yp-url-shortener/internal/logger"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type URLRecord struct {
	URL string
}

type MemStorage struct {
	mu       sync.RWMutex
	storage  map[string]URLRecord
	filePath string
}

func NewMemStorage(filePath string) *MemStorage {
	storage := &MemStorage{storage: make(map[string]URLRecord), filePath: filePath}
	err := storage.LoadFromFile()
	if err != nil {
		logger.Log.Error(err.Error())
	}
	return storage
}

func (s *MemStorage) Save(_ctx context.Context, id string, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage[id] = URLRecord{URL: url}
	err := s.SaveToFile()
	if err != nil {
		logger.Log.Error(err.Error())
		return err
	}
	return nil
}

var ErrRecordNotFound = errors.New("record not found")

func (s *MemStorage) Get(_ctx context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.storage[id]
	if !ok {
		return "", ErrRecordNotFound
	}
	return record.URL, nil
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
