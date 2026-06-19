package hit

import (
	"sync"
	"testing"
)

func TestBufferConcurrent(t *testing.T) {
	b := NewBuffer(100)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Append(Hit{Host: "test", Path: "/"})
		}()
	}
	wg.Wait()
	if b.Len() == 0 && b.Drops() == 0 {
		t.Error("expected hits or drops after concurrent append")
	}
}

func TestBufferFlushEmpty(t *testing.T) {
	b := NewBuffer(16)
	flushed := b.Flush()
	if flushed == nil {
		t.Error("Flush() returned nil, want empty slice")
	}
}

func TestBufferDropsWhenFull(t *testing.T) {
	b := NewBuffer(2)
	b.Append(Hit{Host: "a", Path: "/"})
	b.Append(Hit{Host: "b", Path: "/"})
	b.Append(Hit{Host: "c", Path: "/"})
	if b.Drops() != 1 {
		t.Errorf("Drops() = %d; want 1", b.Drops())
	}
	if b.Len() != 2 {
		t.Errorf("Len() = %d; want 2", b.Len())
	}
}

func TestBufferFlushRace(t *testing.T) {
	b := NewBuffer(200)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Append(Hit{Host: "x", Path: "/"})
		}()
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Flush()
		}()
	}
	wg.Wait()
}
