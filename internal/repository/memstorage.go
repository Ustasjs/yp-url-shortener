package repository

import (
	"errors"
	"sync"
)

type URLRecord struct {
	URL string
}

type MemStorage struct {
	mu      sync.RWMutex
	storage map[string]URLRecord
}

func NewMemStorage() *MemStorage {
	return &MemStorage{storage: make(map[string]URLRecord)}
}

func (s *MemStorage) Save(id string, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage[id] = URLRecord{URL: url}
}

var ErrRecordNotFound = errors.New("record not found")

func (s *MemStorage) Get(id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.storage[id]
	if !ok {
		return "", ErrRecordNotFound
	}
	return record.URL, nil
}
