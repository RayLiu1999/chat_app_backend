package middlewares

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// JSONLoggerMiddleware logs requests in JSON format for Loki
func JSONLoggerMiddleware() gin.HandlerFunc {
	// Setup slog to write JSON to stdout
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		attributes := []any{
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.String("user-agent", c.Request.UserAgent()),
			slog.Duration("latency", latency),
		}

		// Get user ID from context if available (set by auth middleware)
		if userID, exists := c.Get("userID"); exists {
			attributes = append(attributes, slog.Any("user_id", userID))
		}

		if len(c.Errors) > 0 {
			attributes = append(attributes, slog.String("errors", c.Errors.String()))
			logger.Error("Request failed", attributes...)
		} else {
			logger.Info("Request processed", attributes...)
		}
	}
}
