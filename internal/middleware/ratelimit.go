package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Rate limiter manages per-key token buckets.
type RateLimiter struct {
	visitors sync.Map // map[string] *rate.Limiter
	rps      float64
	burst    int
	ttl      time.Duration
}

// New rate limiter creates a limiter with given rate (request/sec ) and burst
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		rps:   rps,
		burst: burst,
		ttl:   5 * time.Minute,
	}
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	if v, ok := rl.visitors.Load(key); ok {
		return v.(*rate.Limiter)
	}
	lim := rate.NewLimiter(rate.Limit(rl.rps), rl.burst)
	rl.visitors.Store(key, lim)
	return lim
}

// Rate limiter enforces per-user or per-IP rate-limit and return middleware for that
// It relies on on GetUser(ctx) to find authenticated user; otherwise it uses remote IP
func RateLimitMiddleware(rl *RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get user from context
			user := GetUser(r.Context())
			var key string
			if user != nil {
				key = user.ID.String()
			} else {
				// fallback use remote IP
				ip, _, _ := net.SplitHostPort(r.RemoteAddr)
				if ip == "" {
					ip = "anon"
				}
				key = "ip:" + ip
			}

			lim := rl.getLimiter(key)
			if !lim.Allow() {
				// Too many request
				w.Header().Set("Retry-After", "1")
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
