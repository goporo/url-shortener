package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple rate limiting middleware
type RateLimiter struct {
	requestsPerMinute int
	clients           map[string][]time.Time
	mu                sync.Mutex
}

// NewRateLimitMiddleware creates a new rate limiter middleware
func NewRateLimitMiddleware(requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		requestsPerMinute: requestsPerMinute,
		clients:           make(map[string][]time.Time),
	}
}

// Limit is the middleware function that limits requests
func (rl *RateLimiter) Limit(c *gin.Context) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Get client IP
	clientIP := c.ClientIP()
	now := time.Now()

	// Clean old requests
	var requests []time.Time
	for _, req := range rl.clients[clientIP] {
		if now.Sub(req) <= time.Minute {
			requests = append(requests, req)
		}
	}

	// Update requests for this client
	rl.clients[clientIP] = requests

	// Check if limit exceeded
	if len(requests) >= rl.requestsPerMinute {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Rate limit exceeded. Try again later.",
		})
		c.Abort()
		return
	}

	// Add current request
	rl.clients[clientIP] = append(rl.clients[clientIP], now)
	c.Next()
}
