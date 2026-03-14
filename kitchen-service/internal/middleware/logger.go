package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// เริ่มจับเวลา
		start := time.Now()

		// ดึง Correlation ID (ที่เราทำไว้ในหัวข้อก่อนหน้า)
		correlationID := c.Get("X-Correlation-ID", "unknown")

		// ให้ Request ทำงานต่อไป
		err := c.Next()

		// หลังทำงานเสร็จ คำนวณเวลาที่ใช้
		latency := time.Since(start)

		// บันทึก Log เป็น JSON
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
