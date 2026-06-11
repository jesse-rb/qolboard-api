package rate_limiting_middleware

import (
	"context"
	"net/http"
	auth_service "qolboard-api/services/auth"
	"qolboard-api/services/logging"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type client struct {
	limiter *rate.Limiter
}

type RateLimiter struct {
	name    string
	mu      sync.Mutex
	clients map[string]*client
}

// PeriodicCleanup starts a background goroutine for periodic cleanups
func (rl *RateLimiter) PeriodicClientsCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop() // Ensure ticker channel is closed on exit

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for key, c := range rl.clients {
				if c.limiter.Allow() {
					delete(rl.clients, key)
				}
			}
			rl.mu.Unlock()
		case <-ctx.Done():
			logging.LogInfo("rate_limiting_middleware", "ctx done, finishing", nil)
			return // Exit when context is canceled
		}
	}
}

func (rl *RateLimiter) RateLimitClient(c *gin.Context, clientKey string, limiter *rate.Limiter) {
	rl.mu.Lock()
	if _, exists := rl.clients[clientKey]; !exists {
		// allow x number of requests per rate, for each clinetKey
		rl.clients[clientKey] = &client{limiter: limiter}
	}
	cl := rl.clients[clientKey]
	rl.mu.Unlock()

	if !cl.limiter.Allow() {
		logging.LogInfo("rate_limiting_middleware", "rate limit exceeded", map[string]any{
			"rate_limiter_name": rl.name,
		})
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": "rate limit exceeded",
		})
	}
}

func RunRateLimitIP(ctx context.Context) gin.HandlerFunc {
	rl := RateLimiter{
		name:    "ip",
		clients: make(map[string]*client),
	}

	go rl.PeriodicClientsCleanup(ctx, 1*time.Minute)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		rl.RateLimitClient(c, ip, rate.NewLimiter(rate.Every(1*time.Second), 10)) // max 10 requests per second for an ip
		c.Next()
	}
}

func RunRateLimitUser(ctx context.Context) gin.HandlerFunc {
	rl := RateLimiter{
		name:    "user",
		clients: make(map[string]*client),
	}

	go rl.PeriodicClientsCleanup(ctx, 1*time.Minute)

	return func(c *gin.Context) {
		id := auth_service.Auth(c)
		rl.RateLimitClient(c, id, rate.NewLimiter(rate.Every(1*time.Second), 10)) // max 10 requests per second for an authenticated user
		c.Next()
	}
}
