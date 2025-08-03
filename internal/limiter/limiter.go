package limiter

import (
	"context"
	"fmt"
	"time"
)

type RateLimiter struct {
	storage Storage
}

func NewRateLimiter(storage Storage) *RateLimiter {
	return &RateLimiter{
		storage: storage,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, limitInfo LimitInfo) (bool, error) {
	blocked, err := rl.storage.IsBlocked(ctx, limitInfo.Key)
	if err != nil {
		return false, fmt.Errorf("failed to check if key is blocked: %w", err)
	}

	if blocked {
		return false, nil
	}

	count, err := rl.storage.Increment(ctx, limitInfo.Key, time.Second)
	if err != nil {
		return false, fmt.Errorf("failed to increment counter: %w", err)
	}

	if count > limitInfo.RateLimit {
		if err := rl.storage.Block(ctx, limitInfo.Key, limitInfo.BlockTime); err != nil {
			return false, fmt.Errorf("failed to block key: %w", err)
		}
		return false, nil
	}

	return true, nil
}

func (rl *RateLimiter) Close() error {
	return rl.storage.Close()
}
