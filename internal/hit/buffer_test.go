package hit

import (
	"sync"
	"testing"
)

func TestBufferConcurrent(t *testing.T) {
	b := NewBuffer()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Append(Hit{Host: "test", Path: "/"})
		}()
	}
	wg.Wait()
	// No race, no panic. Buffer goroutine drains asynchronously.
	// Len reports hits still in the channel; may be 0 if fully drained.
}

func TestBufferFlushEmpty(t *testing.T) {
	b := NewBuffer()
	flushed := b.Flush()
	if flushed == nil {
		t.Error("Flush() returned nil, want empty slice")
	}
}
