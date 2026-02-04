package middleware

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

// RoleMiddleware memeriksa apakah user punya role tertentu
func RoleMiddleware(db *sql.DB, requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(int)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Cannot verify role: user not found in context",
			})
		}

		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM user_roles ur
				JOIN roles r ON ur.role_id = r.id
				WHERE ur.user_id = $1 AND r.name = $2
			)`, userID, requiredRole).Scan(&exists)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Database error",
			})
		}

		if !exists {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied: insufficient permissions",
			})
		}

		return c.Next()
	}
}

// AdminOnly shortcut
func AdminOnly(db *sql.DB) fiber.Handler {
	return RoleMiddleware(db, "admin")
}
