package cache

import (
	"context"
	"sync"
	"time"
)

type CachedItem struct {
	Payload   any
	ExpiresAt time.Time
}

type TTLCache struct {
	mutex        sync.RWMutex
	store        map[string]CachedItem
	maxSize      int
	cleanChannel chan struct{}
}

func NewTTLCache(maxSize int) *TTLCache {
	return &TTLCache{
		store:        make(map[string]CachedItem),
		maxSize:      maxSize,
		cleanChannel: make(chan struct{}),
	}
}

func (cache *TTLCache) StartJanitor(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				cache.EvictExpired()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (cache *TTLCache) Set(key string, value any, duration time.Duration) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// Memory Boundary: Evict arbitrary record if store exceeds maximum size thresholds
	if len(cache.store) >= cache.maxSize && cache.maxSize > 0 {
		for keyToEvict := range cache.store {
			delete(cache.store, keyToEvict)
			break 
		}
	}

	cache.store[key] = CachedItem{
		Payload:   value,
		ExpiresAt: time.Now().Add(duration),
	}
}

func (cache *TTLCache) Get(key string) (any, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	item, exists := cache.store[key]
	if !exists || time.Now().After(item.ExpiresAt) {
		return nil, false
	}
	return item.Payload, true
}

func (cache *TTLCache) EvictExpired() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	now := time.Now()
	for key, item := range cache.store {
		if now.After(item.ExpiresAt) {
			delete(cache.store, key)
		}
	}
}
