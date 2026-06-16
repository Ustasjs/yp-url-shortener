package audit

import "sync"

type Observer interface {
	Notify(event Event)
}

type Notifier struct {
	cond  *sync.Cond
	event Event
}

func NewNotifier() *Notifier {
	return &Notifier{cond: sync.NewCond(&sync.Mutex{})}
}

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

func (n *Notifier) Publish(event Event) {
	n.cond.L.Lock()
	n.event = event
	n.cond.Broadcast()
	n.cond.L.Unlock()
}
