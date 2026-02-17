package middleware

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/pkg/logger"
)

func Logger(log *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			log.Info("request",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", c.Response().Status,
				"latency", time.Since(start).String(),
				"ip", c.RealIP(),
			)

			return err
		}
	}
}
