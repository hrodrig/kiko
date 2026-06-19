package store

import (
	"context"

	"github.com/hrodrig/kiko/internal/hit"
)

type NopStore struct{}

func NewNop() *NopStore { return &NopStore{} }

func (n *NopStore) SaveHits(_ []hit.Hit) error { return nil }

func (n *NopStore) Ping(context.Context) error { return nil }

func (n *NopStore) Close() error { return nil }
