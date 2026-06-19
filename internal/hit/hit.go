package hit

import "sync"

type Hit struct {
	Host        string `json:"host"`
	Path        string `json:"path"`
	Referrer    string `json:"referrer,omitempty"`
	Title       string `json:"title,omitempty"`
	Width       int    `json:"width,omitempty"`
	VisitorHash string `json:"-"`
}

func (h *Hit) Normalize() {
	if h.Path == "" {
		h.Path = "/"
	}
}

type Buffer interface {
	Append(Hit)
	Flush() []Hit
	Len() int
	Drops() uint64
}

type buffer struct {
	mu    sync.Mutex
	hits  []Hit
	cap   int
	drops uint64
}

func NewBuffer(capacity int) Buffer {
	if capacity <= 0 {
		capacity = 4096
	}
	return &buffer{
		hits: make([]Hit, 0, min(capacity, 1024)),
		cap:  capacity,
	}
}

func (b *buffer) Append(h Hit) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.hits) >= b.cap {
		b.drops++
		return
	}
	b.hits = append(b.hits, h)
}

func (b *buffer) Flush() []Hit {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := b.hits
	b.hits = make([]Hit, 0, min(b.cap, 1024))
	return out
}

func (b *buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.hits)
}

func (b *buffer) Drops() uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.drops
}
