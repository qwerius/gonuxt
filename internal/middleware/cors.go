package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS middleware
func CORS() fiber.Handler {
	// Tentukan allowed origins berdasarkan environment
	allowOrigins := "https://blueink.id"
	if os.Getenv("ENV") == "development" || os.Getenv("ENV") == "" {
		allowOrigins = "http://localhost:3000"
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		MaxAge:           300, // Cache preflight 5 menit
	})
}

func SecureHeaders() fiber.Handler {
	env := os.Getenv("ENV")

	return func(c *fiber.Ctx) error {
		c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")

		csp := "default-src * 'unsafe-inline' 'unsafe-eval';"
		if env == "production" {
			csp = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:;"
		}
		c.Set("Content-Security-Policy", csp)

		return c.Next()
	}
}
