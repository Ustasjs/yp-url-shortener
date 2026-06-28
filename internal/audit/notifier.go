package audit

import (
	"io"
	"sync"

	"Ustasjs/yp-url-shortener/internal/logger"

	"go.uber.org/zap"
)

const maxQueueLen = 1000

// Observer receives audit events delivered by a Notifier.
type Observer interface {
	Notify(event Event)
}

// sink couples an observer with its own pending-event queue and the goroutine
// that drains it. Each observer is served independently so a slow one cannot
// block publishing or affect the others.
type sink struct {
	observer Observer

	mu     sync.Mutex
	cond   *sync.Cond
	events []Event
	closed bool
}

// Notifier delivers every published audit event to all attached observers via
// a per-observer queue, implementing a publish/subscribe model. Publishing
// never blocks on slow observers.
type Notifier struct {
	mu    sync.Mutex
	sinks []*sink
	wg    sync.WaitGroup
}

// NewNotifier returns an empty Notifier ready for observers to be attached.
func NewNotifier() *Notifier {
	return &Notifier{}
}

// Attach registers o to receive published events. It starts a background
// goroutine that delivers queued events. Registration is
// synchronous, so an event published immediately after Attach is queued for o
// even before its goroutine starts running.
func (n *Notifier) Attach(o Observer) {
	s := &sink{observer: o}
	s.cond = sync.NewCond(&s.mu)

	n.mu.Lock()
	n.sinks = append(n.sinks, s)
	n.mu.Unlock()

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		for {
			s.mu.Lock()
			for len(s.events) == 0 && !s.closed {
				s.cond.Wait()
			}
			if len(s.events) == 0 && s.closed {
				s.mu.Unlock()
				return
			}
			event := s.events[0]
			s.events = s.events[1:]
			s.mu.Unlock()

			s.observer.Notify(event)
		}
	}()
}

// Publish queues event for every attached observer and wakes their delivery
// goroutines. It never blocks: if an observer's queue is full (its consumer has
// fallen behind), the event is dropped and the drop is logged.
func (n *Notifier) Publish(event Event) {
	n.mu.Lock()
	sinks := n.sinks
	n.mu.Unlock()

	for _, s := range sinks {
		s.mu.Lock()
		if len(s.events) >= maxQueueLen {
			s.mu.Unlock()
			logger.Log.Warn("audit: observer queue full, dropping event", zap.Int("limit", maxQueueLen))
			continue
		}
		s.events = append(s.events, event)
		s.cond.Signal()
		s.mu.Unlock()
	}
}

// Close stops delivery and releases resources. Each observer's goroutine drains
// its remaining queued events before exiting, so already-published events are
// flushed. Observers that implement io.Closer (for example FileObserver) are then closed.
func (n *Notifier) Close() {
	n.mu.Lock()
	sinks := n.sinks
	n.sinks = nil
	n.mu.Unlock()

	for _, s := range sinks {
		s.mu.Lock()
		s.closed = true
		s.cond.Signal()
		s.mu.Unlock()
	}

	n.wg.Wait()

	for _, s := range sinks {
		if c, ok := s.observer.(io.Closer); ok {
			if err := c.Close(); err != nil {
				logger.Log.Error("audit: closing observer failed", zap.Error(err))
			}
		}
	}
}
