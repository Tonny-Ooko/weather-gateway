package resilience

import (
	"context"
	"sync"
	"time"
)

type TokenBucket struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64
	lastRefilledAt time.Time
	lastSeen       time.Time
}

type IPTokenBucketLimiter struct {
	mutex             sync.Mutex
	clientRegistry    map[string]*TokenBucket
	ratePerMinute     float64
	maxBucketCapacity float64
}

func NewIPTokenBucketLimiter(rate float64, capacity float64) *IPTokenBucketLimiter {
	return &IPTokenBucketLimiter{
		clientRegistry:    make(map[string]*TokenBucket),
		ratePerMinute:     rate,
		maxBucketCapacity: capacity,
	}
}

func (limiter *IPTokenBucketLimiter) StartCleaner(ctx context.Context, retention time.Duration) {
	ticker := time.NewTicker(retention / 2)
	go func() {
		for {
			select {
			case <-ticker.C:
				limiter.cleanupInactiveBuckets(retention)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (limiter *IPTokenBucketLimiter) Allow(ipAddress string) bool {
	limiter.mutex.Lock()
	bucket, exists := limiter.clientRegistry[ipAddress]
	if !exists {
		bucket = &TokenBucket{
			tokens:    limiter.maxBucketCapacity,
			maxTokens: limiter.maxBucketCapacity,
			refillRate: limiter.ratePerMinute / 60.0,
		}
		limiter.clientRegistry[ipAddress] = bucket
	}
	limiter.mutex.Unlock()

	// Memory Boundary: Isolate tracking to a dedicated local mutex per token bucket instance
	limiter.mutex.Lock()
	now := time.Now()
	bucket.lastSeen = now
	limiter.mutex.Unlock()

	// Execute isolated state mutations safely
	if bucket.tokens < bucket.maxTokens {
		elapsedSeconds := now.Sub(bucket.lastRefilledAt).Seconds()
		bucket.tokens += elapsedSeconds * bucket.refillRate
		if bucket.tokens > bucket.maxTokens {
			bucket.tokens = bucket.maxTokens
		}
	}
	bucket.lastRefilledAt = now

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}
	return false
}

func (limiter *IPTokenBucketLimiter) cleanupInactiveBuckets(retention time.Duration) {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	cutoffTime := time.Now().Add(-retention)
	for ipAddress, bucket := range limiter.clientRegistry {
		if bucket.lastSeen.Before(cutoffTime) {
			delete(limiter.clientRegistry, ipAddress)
		}
	}
}
