package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// RequestLoggerMiddleware logs every request with zerolog, including method,
// path, status, latency, client IP, and a request ID.
func RequestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Generate or reuse an incoming request ID.
			reqID := c.Request().Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = c.Response().Header().Get(echo.HeaderXRequestID)
			}
			if reqID == "" {
				reqID = generateRequestID()
			}
			c.Response().Header().Set("X-Request-ID", reqID)

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status

			var event *zerolog.Event
			switch {
			case status >= 500:
				event = log.Error()
			case status >= 400:
				event = log.Warn()
			default:
				event = log.Info()
			}

			event.
				Str("request_id", reqID).
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Str("query", c.Request().URL.RawQuery).
				Int("status", status).
				Dur("duration", duration).
				Str("ip", c.RealIP()).
				Str("user_agent", c.Request().UserAgent()).
				Msg("request")

			return err
		}
	}
}

// generateRequestID produces a short unique ID from the current time.
// Good enough for dev/logging; swap with a UUID library in production.
func generateRequestID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	now := time.Now().UnixNano()
	id := make([]byte, 12)
	for i := range id {
		id[i] = chars[now%int64(len(chars))]
		now /= int64(len(chars))
	}
	return string(id)
}
