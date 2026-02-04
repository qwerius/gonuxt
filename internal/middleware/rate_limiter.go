package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

type RateLimitConfig struct {
	Max        int
	Expiration time.Duration
}

func RateLimit(cfg RateLimitConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests",
			})
		},
	})
}
