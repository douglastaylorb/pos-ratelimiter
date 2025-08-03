package storage

import (
	"context"
	"sync"
	"time"
)

type MemoryStorage struct {
	data    map[string]int
	blocks  map[string]time.Time
	expires map[string]time.Time
	mutex   sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		data:    make(map[string]int),
		blocks:  make(map[string]time.Time),
		expires: make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go storage.cleanup()

	return storage
}

func (m *MemoryStorage) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		now := time.Now()

		// Clean expired data
		for key, expiry := range m.expires {
			if now.After(expiry) {
				delete(m.data, key)
				delete(m.expires, key)
			}
		}

		// Clean expired blocks
		for key, expiry := range m.blocks {
			if now.After(expiry) {
				delete(m.blocks, key)
			}
		}

		m.mutex.Unlock()
	}
}

func (m *MemoryStorage) Get(ctx context.Context, key string) (int, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if expiry, exists := m.expires[key]; exists && time.Now().After(expiry) {
		return 0, nil
	}

	value, exists := m.data[key]
	if !exists {
		return 0, nil
	}

	return value, nil
}

func (m *MemoryStorage) Set(ctx context.Context, key string, value int, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data[key] = value
	if expiration > 0 {
		m.expires[key] = time.Now().Add(expiration)
	}

	return nil
}

func (m *MemoryStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if key is expired
	if expiry, exists := m.expires[key]; exists && time.Now().After(expiry) {
		m.data[key] = 0
	}

	m.data[key]++
	if expiration > 0 {
		m.expires[key] = time.Now().Add(expiration)
	}

	return m.data[key], nil
}

func (m *MemoryStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	blockTime, exists := m.blocks[key]
	if !exists {
		return false, nil
	}

	return time.Now().Before(blockTime), nil
}

func (m *MemoryStorage) Block(ctx context.Context, key string, duration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.blocks[key] = time.Now().Add(duration)
	return nil
}

func (m *MemoryStorage) Close() error {
	return nil
}
