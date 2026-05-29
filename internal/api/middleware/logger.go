package middleware

import (
	"fmt"
	"time"

	"Eve/internal/logger"

	"github.com/gofiber/fiber/v3"
)

func RequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		status := c.Response().StatusCode()
		logger.HTTP(
			fmt.Sprintf("%d %s %s", status, c.Method(), c.Path()),
			"latency", latency.Round(time.Microsecond),
			"ip", c.IP(),
		)
		return err
	}
}
