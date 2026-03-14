package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// เริ่มจับเวลา
		start := time.Now()

		// ดึง Correlation ID จาก Gateway หรือสร้างใหม่ถ้าไม่มี
		correlationID := c.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
			c.Set("X-Correlation-ID", correlationID) // ส่งกลับใน response header ด้วย
		}
		// เก็บไว้ใน Locals เพื่อให้ handler ทุกตัวดึงได้ง่าย
		c.Locals("correlationID", correlationID)

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
