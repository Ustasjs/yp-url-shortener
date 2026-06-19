package audit

import (
	"Ustasjs/yp-url-shortener/internal/logger"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

// FileObserver is an audit Observer that appends events as JSON lines to a file.
// It is safe for concurrent use.
type FileObserver struct {
	mu   sync.Mutex
	file *os.File
}

// NewFileObserver opens (creating if necessary) the file at path for appending
// and returns a FileObserver that writes to it.
func NewFileObserver(path string) (*FileObserver, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}
	return &FileObserver{file: file}, nil
}

// Notify appends event to the file as a single JSON line. Errors are logged
// rather than returned.
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

// Close closes the underlying file.
func (f *FileObserver) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Close()
}
