package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		correlationID := c.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
			c.Set("X-Correlation-ID", correlationID)
		}
		c.Locals("correlationID", correlationID)

		err := c.Next()

		latency := time.Since(start)

		event := log.Info()
		if err != nil {
			event = log.Error().Err(err)
		}

		event.
			Str("correlation_id", correlationID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Str("latency", latency.String()).
			Str("ip", c.IP()).
			Msg("HTTP Request")

		return err
	}
}
