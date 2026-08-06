package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

func NewIPRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	entries := make(map[string]rateLimitEntry)
	return func(c *gin.Context) {
		now := time.Now()
		key := c.ClientIP()
		mu.Lock()
		entry := entries[key]
		if entry.resetAt.IsZero() || !now.Before(entry.resetAt) {
			entry = rateLimitEntry{resetAt: now.Add(window)}
		}
		if entry.count >= limit {
			retryAfter := int(time.Until(entry.resetAt).Seconds()) + 1
			mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		entry.count++
		entries[key] = entry
		if len(entries) > 10_000 {
			for ip, candidate := range entries {
				if !now.Before(candidate.resetAt) {
					delete(entries, ip)
				}
			}
		}
		mu.Unlock()
		c.Next()
	}
}
