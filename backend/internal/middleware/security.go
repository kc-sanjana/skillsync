package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// SecurityHeadersMiddleware sets common security-related HTTP headers.
func SecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-XSS-Protection", "1; mode=block")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			return next(c)
		}
	}
}

// RequestSizeLimitMiddleware rejects request bodies larger than the given
// number of bytes.
func RequestSizeLimitMiddleware(maxBytes int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().ContentLength > maxBytes {
				return c.JSON(http.StatusRequestEntityTooLarge, map[string]string{
					"error": "request body too large",
				})
			}
			c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxBytes)
			return next(c)
		}
	}
}

// ---------------------------------------------------------------------------
// In-memory sliding-window rate limiter (per IP)
// ---------------------------------------------------------------------------

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

type visitor struct {
	timestamps []time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	// Background cleanup every minute.
	go func() {
		for {
			time.Sleep(time.Minute)
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{}
		rl.visitors[ip] = v
	}

	// Prune expired timestamps.
	valid := v.timestamps[:0]
	for _, t := range v.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	v.timestamps = valid

	if len(v.timestamps) >= rl.limit {
		return false
	}

	v.timestamps = append(v.timestamps, now)
	return true
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.window)
	for ip, v := range rl.visitors {
		valid := v.timestamps[:0]
		for _, t := range v.timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.visitors, ip)
		} else {
			v.timestamps = valid
		}
	}
}

// RateLimitMiddleware limits each IP to `limit` requests per `window`.
// Default: 100 requests per minute.
func RateLimitMiddleware(limit int, window time.Duration) echo.MiddlewareFunc {
	rl := newRateLimiter(limit, window)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			if !rl.allow(ip) {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "rate limit exceeded, try again later",
				})
			}
			return next(c)
		}
	}
}
