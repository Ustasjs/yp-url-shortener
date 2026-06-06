package audit_test

import (
	"Ustasjs/yp-url-shortener/internal/audit"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type recorder struct {
	mu     sync.Mutex
	events []audit.Event
	delay  time.Duration
}

func (r *recorder) Notify(event audit.Event) {
	if r.delay > 0 {
		time.Sleep(r.delay)
	}
	r.mu.Lock()
	r.events = append(r.events, event)
	r.mu.Unlock()
}

func (r *recorder) snapshot() []audit.Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]audit.Event, len(r.events))
	copy(out, r.events)
	return out
}

func TestNotifier_BroadcastsToAllObservers(t *testing.T) {
	notifier := audit.NewNotifier()
	a, b := &recorder{}, &recorder{}
	notifier.Attach(a)
	notifier.Attach(b)

	time.Sleep(50 * time.Millisecond)

	event := audit.Event{Timestamp: 1, Action: audit.ActionShorten, UserID: "u1", URL: "https://x"}
	notifier.Publish(event)

	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, []audit.Event{event}, a.snapshot())
	assert.Equal(t, []audit.Event{event}, b.snapshot())
}

func TestNotifier_PublishIsNonBlocking(t *testing.T) {
	notifier := audit.NewNotifier()
	notifier.Attach(&recorder{delay: 200 * time.Millisecond})
	time.Sleep(50 * time.Millisecond)

	start := time.Now()
	notifier.Publish(audit.Event{Action: audit.ActionFollow})
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 20*time.Millisecond)
}

func TestNotifier_PublishWithoutObservers(t *testing.T) {
	notifier := audit.NewNotifier()
	notifier.Publish(audit.Event{Action: audit.ActionShorten})
}
