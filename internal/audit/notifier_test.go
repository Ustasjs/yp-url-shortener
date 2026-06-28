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

func TestNotifier_NoLossWithSlowObserver(t *testing.T) {
	notifier := audit.NewNotifier()
	r := &recorder{delay: 10 * time.Millisecond}
	notifier.Attach(r)

	want := make([]audit.Event, 5)
	for i := range want {
		want[i] = audit.Event{Timestamp: int64(i), Action: audit.ActionShorten}
		notifier.Publish(want[i])
	}

	assert.Eventually(t, func() bool {
		return len(r.snapshot()) == len(want)
	}, time.Second, 10*time.Millisecond)
	assert.Equal(t, want, r.snapshot())
}

func TestNotifier_DropsWhenQueueFull(t *testing.T) {
	notifier := audit.NewNotifier()
	// A long delay keeps the consumer busy so the queue fills past the cap.
	notifier.Attach(&recorder{delay: time.Hour})

	const published = 1500 // > maxQueueLen (1000)
	start := time.Now()
	for i := 0; i < published; i++ {
		notifier.Publish(audit.Event{Timestamp: int64(i), Action: audit.ActionShorten})
	}

	// Publishing must stay non-blocking even though the queue is saturated.
	assert.Less(t, time.Since(start), time.Second)
}

func TestNotifier_CloseFlushesAndStops(t *testing.T) {
	notifier := audit.NewNotifier()
	r := &recorder{}
	notifier.Attach(r)

	event := audit.Event{Timestamp: 1, Action: audit.ActionShorten}
	notifier.Publish(event)

	notifier.Close()

	// Close drains queued events before stopping the goroutine.
	assert.Equal(t, []audit.Event{event}, r.snapshot())

	// Close is idempotent.
	assert.NotPanics(t, func() { notifier.Close() })
}
