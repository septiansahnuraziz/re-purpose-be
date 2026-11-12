package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/utils/httpresponse"
)

type rateLimitEntry struct {
	count int
	reset time.Time
}

// RateLimiter is a per-client fixed-window limiter. It is in-memory and
// per-process; use a shared store (e.g. Redis) when running multiple replicas.
type RateLimiter struct {
	mu          sync.Mutex
	limit       int
	window      time.Duration
	clients     map[string]*rateLimitEntry
	lastCleanup time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 30
	}
	if window <= 0 {
		window = time.Minute
	}

	return &RateLimiter{
		limit:       limit,
		window:      window,
		clients:     make(map[string]*rateLimitEntry),
		lastCleanup: time.Now(),
	}
}

// Middleware rejects requests once a client exceeds the configured limit and
// responds with the standard RPUR-429001 envelope.
func (l *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.allow(c.ClientIP(), time.Now()) {
			httpresponse.Error(c, &httpresponse.ErrorTooManyRequests)
			return
		}
		c.Next()
	}
}

func (l *RateLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cleanupLocked(now)

	entry, ok := l.clients[key]
	if !ok || now.After(entry.reset) {
		l.clients[key] = &rateLimitEntry{count: 1, reset: now.Add(l.window)}
		return true
	}
	if entry.count >= l.limit {
		return false
	}

	entry.count++
	return true
}

func (l *RateLimiter) cleanupLocked(now time.Time) {
	if now.Sub(l.lastCleanup) < l.window {
		return
	}

	for key, entry := range l.clients {
		if now.After(entry.reset) {
			delete(l.clients, key)
		}
	}
	l.lastCleanup = now
}
