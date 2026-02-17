package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

// CORSMiddleware returns Echo middleware that sets CORS headers based on
// the CORS_ALLOWED_ORIGINS env var (comma-separated). Falls back to
// localhost origins for development.
func CORSMiddleware() echo.MiddlewareFunc {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		raw = "http://localhost:3000,http://localhost:5173"
	}
	allowed := strings.Split(raw, ",")
	originSet := make(map[string]bool, len(allowed))
	for _, o := range allowed {
		originSet[strings.TrimSpace(o)] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			origin := c.Request().Header.Get("Origin")

			if originSet[origin] || originSet["*"] {
				c.Response().Header().Set("Access-Control-Allow-Origin", origin)
			}

			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
			c.Response().Header().Set("Access-Control-Allow-Methods",
				"GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers",
				"Origin, Content-Type, Accept, Authorization, X-Request-ID")
			c.Response().Header().Set("Access-Control-Expose-Headers",
				"X-Request-ID")
			c.Response().Header().Set("Access-Control-Max-Age", "86400")

			if c.Request().Method == http.MethodOptions {
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	}
}
