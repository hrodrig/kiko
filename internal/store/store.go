package store

import "github.com/hrodrig/kiko/internal/hit"

type Store interface {
	SaveHits(hits []hit.Hit) error
}

type NopStore struct{}

func NewNop() *NopStore                        { return &NopStore{} }
func (n *NopStore) SaveHits(_ []hit.Hit) error { return nil }
