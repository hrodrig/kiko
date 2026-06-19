package cli

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/hrodrig/kiko/internal/config"
	"github.com/hrodrig/kiko/internal/hit"
	"github.com/hrodrig/kiko/internal/log"
)

type fakeStore struct {
	mu    sync.Mutex
	saved int
	err   error
	ping  error
}

func (f *fakeStore) SaveHits(hits []hit.Hit) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.saved += len(hits)
	return nil
}

func (f *fakeStore) savedCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.saved
}

func (f *fakeStore) Ping(context.Context) error { return f.ping }

func (f *fakeStore) Close() error { return nil }

func TestFlushHitsEmpty(t *testing.T) {
	st := &fakeStore{}
	buf := hit.NewBuffer(16)
	flushHits(st, buf, log.New(nil, log.Off))
	if st.savedCount() != 0 {
		t.Errorf("saved = %d; want 0", st.savedCount())
	}
}

func TestFlushHitsPersists(t *testing.T) {
	st := &fakeStore{}
	buf := hit.NewBuffer(16)
	buf.Append(hit.Hit{Host: "a", Path: "/"})
	buf.Append(hit.Hit{Host: "b", Path: "/x"})
	flushHits(st, buf, log.New(nil, log.Off))
	if st.savedCount() != 2 {
		t.Errorf("saved = %d; want 2", st.savedCount())
	}
	if buf.Len() != 0 {
		t.Errorf("Len after flush = %d; want 0", buf.Len())
	}
}

func TestFlushHitsError(t *testing.T) {
	st := &fakeStore{err: errors.New("db down")}
	buf := hit.NewBuffer(16)
	buf.Append(hit.Hit{Host: "a", Path: "/"})
	flushHits(st, buf, log.New(nil, log.Off))
	if st.savedCount() != 0 {
		t.Errorf("saved = %d; want 0 on error", st.savedCount())
	}
}

func TestRunFlusher(t *testing.T) {
	st := &fakeStore{}
	buf := hit.NewBuffer(16)
	cfg := &config.Config{
		Buffer: config.BufferCfg{FlushInterval: 1},
		Log:    log.New(nil, log.Off),
	}
	ctx, cancel := context.WithCancel(context.Background())

	go runFlusher(ctx, st, buf, cfg)
	buf.Append(hit.Hit{Host: "a", Path: "/"})
	time.Sleep(1200 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)

	if st.savedCount() < 1 {
		t.Errorf("saved = %d; want >= 1", st.savedCount())
	}
}
