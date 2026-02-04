package middleware

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// OwnerOrAdminMiddleware memastikan user hanya bisa mengubah profile miliknya sendiri, kecuali admin
func OwnerOrAdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDLocal := c.Locals("user_id")
		roleLocal := c.Locals("role")

		if userIDLocal == nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		// konversi user_id ke int
		var userIDInt int
		switch v := userIDLocal.(type) {
		case int:
			userIDInt = v
		case int64:
			userIDInt = int(v)
		case float64:
			userIDInt = int(v)
		case string:
			id, err := strconv.Atoi(v)
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "invalid user_id",
				})
			}
			userIDInt = id
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "invalid user_id type",
			})
		}

		// debug: lihat user_id dan role
		log.Println("DEBUG OwnerOrAdminMiddleware user_id:", userIDInt, "role:", roleLocal)

		// cek role admin
		if roleStr, ok := roleLocal.(string); ok && roleStr == "admin" {
			// admin selalu boleh
			return c.Next()
		}

		// ambil id dari param
		paramID := c.Params("id")
		if paramID == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "missing id param",
			})
		}

		paramIDInt, err := strconv.Atoi(paramID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid id param",
			})
		}

		// cek apakah user id sama dengan param id
		if userIDInt != paramIDInt {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden: cannot access other user's profile",
			})
		}

		return c.Next()
	}
}
