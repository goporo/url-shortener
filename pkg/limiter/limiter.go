package limiter

import (
	"sync"
	"time"
)

// RateLimiter defines a token bucket rate limiter
type RateLimiter struct {
	mu           sync.Mutex
	tokens       map[string]*bucket
	rate         int           // tokens per interval
	interval     time.Duration // refill interval
	maxTokens    int           // maximum tokens per bucket
	cleanupAfter time.Duration // how long to keep buckets in memory
}

type bucket struct {
	tokens    int
	lastSeen  time.Time
	lastRefil time.Time
}

func NewRateLimiter(rate int, interval time.Duration, maxTokens int) *RateLimiter {
	limiter := &RateLimiter{
		tokens:       make(map[string]*bucket),
		rate:         rate,
		interval:     interval,
		maxTokens:    maxTokens,
		cleanupAfter: time.Hour,
	}

	go limiter.cleanup()

	return limiter
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.tokens[key]

	if !exists {
		rl.tokens[key] = &bucket{
			tokens:    rl.maxTokens - 1,
			lastSeen:  now,
			lastRefil: now,
		}
		return true
	}

	b.lastSeen = now

	elapsed := now.Sub(b.lastRefil)
	tokensToAdd := int(elapsed/rl.interval) * rl.rate

	if tokensToAdd > 0 {
		b.tokens = min(b.tokens+tokensToAdd, rl.maxTokens)
		b.lastRefil = now
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) RemainingTokens(key string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.tokens[key]
	if !exists {
		return rl.maxTokens
	}

	now := time.Now()
	elapsed := now.Sub(b.lastRefil)
	tokensToAdd := int(elapsed/rl.interval) * rl.rate

	return min(b.tokens+tokensToAdd, rl.maxTokens)
}

func (rl *RateLimiter) NextAvailable(key string) time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.tokens[key]
	if !exists || b.tokens > 0 {
		return 0
	}

	now := time.Now()
	elapsed := now.Sub(b.lastRefil)
	remaining := rl.interval - (elapsed % rl.interval)

	return remaining
}

// cleanup periodically removes inactive buckets to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.tokens {
			if now.Sub(bucket.lastSeen) > rl.cleanupAfter {
				delete(rl.tokens, key)
			}
		}
		rl.mu.Unlock()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
