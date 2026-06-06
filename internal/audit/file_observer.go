package audit

import (
	"Ustasjs/yp-url-shortener/internal/logger"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

type FileObserver struct {
	mu   sync.Mutex
	file *os.File
}

func NewFileObserver(path string) (*FileObserver, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}
	return &FileObserver{file: file}, nil
}

func (f *FileObserver) Notify(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error("audit: marshal event failed", zap.Error(err))
		return
	}
	data = append(data, '\n')

	f.mu.Lock()
	defer f.mu.Unlock()
	if _, err := f.file.Write(data); err != nil {
		logger.Log.Error("audit: write to file failed", zap.Error(err))
	}
}

func (f *FileObserver) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Close()
}
