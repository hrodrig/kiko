package hit

type Hit struct {
	Host     string `json:"host"`
	Path     string `json:"path"`
	Referrer string `json:"referrer,omitempty"`
	Title    string `json:"title,omitempty"`
	Width    int    `json:"width,omitempty"`
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
}

type buffer struct {
	hits []Hit
	ch   chan Hit
}

func NewBuffer() Buffer {
	b := &buffer{
		hits: make([]Hit, 0, 1024),
		ch:   make(chan Hit, 4096),
	}
	go func() {
		for h := range b.ch {
			b.hits = append(b.hits, h)
		}
	}()
	return b
}

func (b *buffer) Append(h Hit) {
	select {
	case b.ch <- h:
	default:
		// drop if channel full
	}
}

func (b *buffer) Flush() []Hit {
	out := b.hits
	b.hits = make([]Hit, 0, 1024)
	return out
}

func (b *buffer) Len() int {
	return len(b.hits)
}
