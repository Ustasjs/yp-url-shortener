// Package pool provides a generic, type-safe wrapper around sync.Pool for
// reusing heavy objects. Every object is reset to its zero state before it is
// returned to the pool, which pairs naturally with the Reset() methods produced
// by cmd/reset.
package pool

import "sync"

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	pool sync.Pool
}

// New creates a Pool that allocates fresh values with newFn whenever the pool is
// empty. newFn is optional: when it is nil, Get returns the zero value of T for
// an empty pool instead of a constructed object.
func New[T Resetter](newFn func() T) *Pool[T] {
	p := &Pool[T]{}
	if newFn != nil {
		p.pool.New = func() any { return newFn() }
	}
	return p
}

// Get returns an object from the pool. When the pool is empty it allocates a new
// one via the factory passed to New, or returns the zero value of T if no factory
// was provided.
func (p *Pool[T]) Get() T {
	v := p.pool.Get()
	if v == nil {
		var zero T
		return zero
	}
	return v.(T)
}

// Put resets the object to its zero state and returns it to the pool so it can
// be reused by a later Get.
func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
