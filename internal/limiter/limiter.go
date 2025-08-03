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
	// Check if the key is currently blocked
	blocked, err := rl.storage.IsBlocked(ctx, limitInfo.Key)
	if err != nil {
		return false, fmt.Errorf("failed to check if key is blocked: %w", err)
	}

	if blocked {
		return false, nil
	}

	// Increment the counter for the key
	count, err := rl.storage.Increment(ctx, limitInfo.Key, time.Second)
	if err != nil {
		return false, fmt.Errorf("failed to increment counter: %w", err)
	}

	// Check if the limit has been exceeded
	if count > limitInfo.RateLimit {
		// Block the key for the specified duration
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
