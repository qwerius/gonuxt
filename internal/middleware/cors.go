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