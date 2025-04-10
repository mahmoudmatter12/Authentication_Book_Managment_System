package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter holds rate limiting information for each IP
type RateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitMiddleware provides rate limiting per IP address
func RateLimitMiddleware(rps int, burst int, duration time.Duration) gin.HandlerFunc {
	var (
		clients = make(map[string]*RateLimiter)
		mu      sync.Mutex
	)

	// Clean up old entries periodically
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 2*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		defer mu.Unlock()

		// Initialize rate limiter for new IPs
		if _, found := clients[ip]; !found {
			clients[ip] = &RateLimiter{
				limiter: rate.NewLimiter(rate.Limit(rps), burst),
			}
		}

		// Update last seen time
		clients[ip].lastSeen = time.Now()

		// Check if request is allowed
		if !clients[ip].limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": duration.String(),
			})
			return
		}

		c.Next()
	}
}