package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/utils"
)

func AuthRequired(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization") // ambil header Authorization
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Authorization header is required",
		})
	}

	// format "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Authorization header format must be Bearer {token}",
		})
	}

	userID, err := utils.ValidateAccessToken(parts[1])
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid token",
		})
	}

	// simpan userID di locals supaya handler bisa pakai
	c.Locals("user_id", userID)
	return c.Next()
}
