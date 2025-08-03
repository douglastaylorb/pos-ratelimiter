package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(host, port, password string, db int) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{client: rdb}, nil
}

func (r *RedisStorage) Get(ctx context.Context, key string) (int, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *RedisStorage) Set(ctx context.Context, key string, value int, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int, error) {
	pipe := r.client.TxPipeline()

	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return int(incr.Val()), nil
}

func (r *RedisStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	blockKey := fmt.Sprintf("block:%s", key)
	exists, err := r.client.Exists(ctx, blockKey).Result()
	return exists > 0, err
}

func (r *RedisStorage) Block(ctx context.Context, key string, duration time.Duration) error {
	blockKey := fmt.Sprintf("block:%s", key)
	return r.client.Set(ctx, blockKey, "blocked", duration).Err()
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}
