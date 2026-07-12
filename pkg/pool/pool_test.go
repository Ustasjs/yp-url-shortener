package pool_test

import (
	"testing"

	"Ustasjs/yp-url-shortener/pkg/pool"
)

// buffer is a minimal Resetter used to exercise the pool.
type buffer struct {
	data []byte
	used bool
}

func (b *buffer) Reset() {
	if b == nil {
		return
	}
	b.data = b.data[:0]
	b.used = false
}

func TestGetAllocatesWhenEmpty(t *testing.T) {
	calls := 0
	p := pool.New(func() *buffer {
		calls++
		return &buffer{}
	})

	got := p.Get()
	if got == nil {
		t.Fatal("Get returned nil")
	}
	if calls != 1 {
		t.Fatalf("factory calls = %d, want 1", calls)
	}
}

func TestGetReturnsZeroValueWhenNoFactory(t *testing.T) {
	p := pool.New[*buffer](nil)

	got := p.Get()
	if got != nil {
		t.Errorf("Get returned %v, want zero value (nil)", got)
	}
}

func TestPutResetsBeforeReuse(t *testing.T) {
	p := pool.New(func() *buffer { return &buffer{} })

	b := p.Get()
	b.data = append(b.data, "payload"...)
	b.used = true

	p.Put(b)

	// sync.Pool does not guarantee the same object is returned, so only assert
	// state when we get our object back.
	reused := p.Get()
	if reused == b {
		if reused.used {
			t.Error("used flag was not reset")
		}
		if len(reused.data) != 0 {
			t.Errorf("data length = %d, want 0", len(reused.data))
		}
		if cap(reused.data) == 0 {
			t.Error("backing array was discarded; slice should be truncated, not niled")
		}
	}
}
