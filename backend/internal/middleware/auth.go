package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/pkg/auth"
)

const userIDKey = "user_id"

// JWTMiddleware returns Echo middleware that validates a Bearer token from the
// Authorization header and stores the authenticated user_id in the context.
func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing authorization header",
				})
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid authorization format, expected: Bearer <token>",
				})
			}

			claims, err := auth.ValidateToken(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid or expired token",
				})
			}

			c.Set(userIDKey, claims.UserID)
			return next(c)
		}
	}
}

// ExtractUserID pulls the authenticated user's ID from the Echo context.
// Must be called from a handler that sits behind JWTMiddleware.
func ExtractUserID(c echo.Context) (string, error) {
	id, ok := c.Get(userIDKey).(string)
	if !ok {
		return "", errors.New("user_id not found in context")
	}
	return id, nil
}
