package limiter

import (
	"context"
	"time"
)

type Storage interface {
	Get(ctx context.Context, key string) (int, error)
	Set(ctx context.Context, key string, value int, expiration time.Duration) error
	Increment(ctx context.Context, key string, expiration time.Duration) (int, error)
	IsBlocked(ctx context.Context, key string) (bool, error)
	Block(ctx context.Context, key string, duration time.Duration) error
	Close() error
}

type LimitInfo struct {
	Key       string
	RateLimit int
	BlockTime time.Duration
}
