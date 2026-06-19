package audit

import "sync"

// Observer receives audit events delivered by a Notifier.
type Observer interface {
	Notify(event Event)
}

// Notifier broadcasts the most recent audit event to all attached observers,
// implementing a simple publish/subscribe model on top of sync.Cond.
type Notifier struct {
	cond  *sync.Cond
	event Event
}

// NewNotifier returns an empty Notifier ready for observers to be attached.
func NewNotifier() *Notifier {
	return &Notifier{cond: sync.NewCond(&sync.Mutex{})}
}

// Attach registers o to receive published events. It starts a background
// goroutine that delivers each broadcast event to o.
func (n *Notifier) Attach(o Observer) {
	go func() {
		for {
			n.cond.L.Lock()
			n.cond.Wait()
			event := n.event
			n.cond.L.Unlock()

			o.Notify(event)
		}
	}()
}

// Publish stores latest event and wakes every attached observer.
func (n *Notifier) Publish(event Event) {
	n.cond.L.Lock()
	n.event = event
	n.cond.Broadcast()
	n.cond.L.Unlock()
}
