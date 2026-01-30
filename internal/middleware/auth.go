package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/utils"
)

func AuthRequired(c *fiber.Ctx) error {
	// 1. coba ambil dari header
	authHeader := c.Get("Authorization")

	var token string

	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Authorization header format must be Bearer {token}",
			})
		}
		token = parts[1]
	} else {
		// 2. kalau header kosong, ambil dari cookie
		token = c.Cookies("access_token")
	}

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "token is required",
		})
	}

	userID, err := utils.ValidateAccessToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid token",
		})
	}

	c.Locals("user_id", userID)
	return c.Next()
}
