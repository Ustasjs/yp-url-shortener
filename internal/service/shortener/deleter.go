package shortener

import (
	"context"
	"errors"
	"sync"
	"time"

	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/model"

	"go.uber.org/zap"
)

var ErrServiceOverloaded = errors.New("service is overloaded")

type deleteTask struct {
	userID   string
	shortIDs []string
}

type URLDeleter struct {
	repo          Storage
	inputChan     chan deleteTask
	batchSize     int
	flushInterval time.Duration

	workersWG   sync.WaitGroup
	batcherDone chan struct{}
	stopOnce    sync.Once
}

func NewURLDeleter(repo Storage, batchSize int, flushInterval time.Duration) *URLDeleter {
	return &URLDeleter{
		repo:          repo,
		inputChan:     make(chan deleteTask, 100),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		batcherDone:   make(chan struct{}),
	}
}

func (d *URLDeleter) Start(numWorkers int) {
	fanInChan := make(chan model.DeleteItem, 1000)

	for i := 0; i < numWorkers; i++ {
		d.workersWG.Add(1)
		go d.worker(fanInChan)
	}

	go d.batcher(fanInChan)

	// Once every worker has drained inputChan and exited, close fanInChan so the
	// batcher can flush its remaining buffer and stop.
	go func() {
		d.workersWG.Wait()
		close(fanInChan)
	}()
}

// Stop gracefully drains the delete pipeline and blocks until every pending
// deletion has been flushed to storage. It must be called only after all HTTP
// handlers have stopped, so no new task can be enqueued onto the closed
// inputChan.
func (d *URLDeleter) Stop() {
	d.stopOnce.Do(func() {
		close(d.inputChan)
	})
	<-d.batcherDone
}

func (d *URLDeleter) worker(out chan<- model.DeleteItem) {
	defer d.workersWG.Done()
	for task := range d.inputChan {
		for _, id := range task.shortIDs {
			out <- model.DeleteItem{UserID: task.userID, ShortID: id}
		}
	}
}

func (d *URLDeleter) batcher(in <-chan model.DeleteItem) {
	defer close(d.batcherDone)

	buffer := make([]model.DeleteItem, 0, d.batchSize)
	ticker := time.NewTicker(d.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case item, ok := <-in:
			if !ok {
				if len(buffer) > 0 {
					d.flush(buffer)
				}
				return
			}
			buffer = append(buffer, item)
			if len(buffer) >= d.batchSize {
				d.flush(buffer)
				buffer = buffer[:0]
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				d.flush(buffer)
				buffer = buffer[:0]
			}
		}
	}
}

func (d *URLDeleter) flush(items []model.DeleteItem) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := d.repo.DeleteURLsBatch(ctx, items)
	if err != nil {
		logger.Log.Error("Failed to delete URLs batch", zap.Error(err), zap.Int("count", len(items)))
	} else {
		logger.Log.Info("Successfully deleted URLs batch", zap.Int("count", len(items)))
	}
}

func (d *URLDeleter) DeleteURLsAsync(userID string, shortIDs []string) error {
	if len(shortIDs) == 0 {
		return nil
	}

	select {
	case d.inputChan <- deleteTask{userID: userID, shortIDs: shortIDs}:
		return nil
	default:
		logger.Log.Warn("Delete channel is full, service overloaded",
			zap.String("userID", userID),
			zap.Int("shortIDsCount", len(shortIDs)))
		return ErrServiceOverloaded
	}
}
