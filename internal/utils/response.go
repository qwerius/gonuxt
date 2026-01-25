package utils

import "github.com/gofiber/fiber/v2"

// Success mengembalikan response status 200 dengan data
func Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "ok",
		"data":   data,
	})
}

// Error mengembalikan response dengan status error dan pesan
func Error(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{
		"status": "error",
		"error":  msg,
	})
}
