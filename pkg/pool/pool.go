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
// empty. newFn must return a non-nil, usable value of type T.
func New[T Resetter](newFn func() T) *Pool[T] {
	p := &Pool[T]{}
	p.pool.New = func() any { return newFn() }
	return p
}

// Get returns an object from the pool, allocating a new one via the factory
// passed to New when the pool is empty.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put resets the object to its zero state and returns it to the pool so it can
// be reused by a later Get.
func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
