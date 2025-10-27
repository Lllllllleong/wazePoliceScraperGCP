package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// IPRateLimiter tracks request rates per IP address
type IPRateLimiter struct {
	ips map[string]*rateLimiter
	mu  *sync.RWMutex
	r   int           // requests
	d   time.Duration // per duration
}

type rateLimiter struct {
	tokens         int
	lastRefillTime time.Time
	mu             *sync.Mutex
}

// NewIPRateLimiter creates a new IP-based rate limiter
// Example: NewIPRateLimiter(10, time.Minute) = 10 requests per minute per IP
func NewIPRateLimiter(r int, d time.Duration) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rateLimiter),
		mu:  &sync.RWMutex{},
		r:   r,
		d:   d,
	}

	// Cleanup old entries every 10 minutes
	go i.cleanupRoutine()

	return i
}

// Allow checks if the IP is allowed to make a request
func (i *IPRateLimiter) Allow(ip string) bool {
	i.mu.Lock()
	limiter, exists := i.ips[ip]

	if !exists {
		limiter = &rateLimiter{
			tokens:         i.r,
			lastRefillTime: time.Now(),
			mu:             &sync.Mutex{},
		}
		i.ips[ip] = limiter
	}
	i.mu.Unlock()

	return limiter.allow(i.r, i.d)
}

func (rl *rateLimiter) allow(rate int, duration time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefillTime)

	// Refill tokens based on time elapsed
	if elapsed >= duration {
		rl.tokens = rate
		rl.lastRefillTime = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// cleanupRoutine removes old IP entries to prevent memory leaks
func (i *IPRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		for ip, limiter := range i.ips {
			limiter.mu.Lock()
			if time.Since(limiter.lastRefillTime) > 30*time.Minute {
				delete(i.ips, ip)
			}
			limiter.mu.Unlock()
		}
		i.mu.Unlock()
	}
}

// RateLimitMiddleware returns a middleware that limits requests per IP
func RateLimitMiddleware(limiter *IPRateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := getIPAddress(r)

			if !limiter.Allow(ip) {
				log.Printf("⚠️ Rate limit exceeded for IP: %s", ip)
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next(w, r)
		}
	}
}

// getIPAddress extracts the real IP address from the request
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
