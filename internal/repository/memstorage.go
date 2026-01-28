package repository

import "errors"

type URLRecord struct {
	URL string
}

type MemStorage struct {
	storage map[string]URLRecord
}

func NewMemStorage() *MemStorage {
	return &MemStorage{storage: make(map[string]URLRecord)}
}

func (s *MemStorage) Save(id string, url string) {
	s.storage[id] = URLRecord{URL: url}
}

func (s *MemStorage) Get(id string) (string, error) {
	record, ok := s.storage[id]
	if !ok {
		return "", errors.New("record not found")
	}
	return record.URL, nil
}
