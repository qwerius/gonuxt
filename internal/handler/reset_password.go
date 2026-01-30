package handler

import (
	"database/sql"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func ResetPassword(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type Request struct {
			Token       string `json:"token"`
			NewPassword string `json:"new_password"`
		}

		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}

		// 1️⃣ validasi token JWT
		userID, err := utils.ValidateAccessToken(req.Token)
		if err != nil {
			log.Println("JWT validation error:", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		// 2️⃣ hash password baru
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Password hash error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
		}

		// 3️⃣ update password di DB
		query := "UPDATE users SET password = $1, updated_at = $2 WHERE id = $3"
		_, err = db.Exec(query, hashedPassword, time.Now(), userID)
		if err != nil {
			log.Println("DB update error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update password"})
		}

		// 4️⃣ respon sukses
		return c.JSON(fiber.Map{"message": "Password berhasil diubah"})
	}
}
