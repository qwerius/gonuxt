package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/smtp"


	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/utils"
	"github.com/qwerius/gonuxt/internal/config"
)

func ForgotPassword(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type Request struct {
			Email string `json:"email"`
		}
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}

		// cek email ada di DB
		var userID int
		query := "SELECT id FROM users WHERE email = $1"
		err := db.QueryRow(query, req.Email).Scan(&userID)
		if err != nil {
			if err == sql.ErrNoRows {
				// email tidak ada
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Email tidak terdaftar",
				})
			}
			log.Println("DB error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}

		// generate JWT token pakai utils
		tokenString, err := utils.CreateAccessToken(userID)
		if err != nil {
			log.Println("JWT generation error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}

		// build reset URL
		frontendURL := config.Get("FRONTEND_URL")
		resetURL := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, tokenString)

		// kirim email nyata via SMTP
		if err := sendResetEmailSMTP(req.Email, resetURL); err != nil {
			log.Println("Email send error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to send email"})
		}

		return c.JSON(fiber.Map{"message": "Link reset password telah dikirim"})
	}
}

// sendResetEmailSMTP mengirim email via SMTP sesuai env
func sendResetEmailSMTP(to, resetURL string) error {
	host := config.Get("EMAIL_HOST")
	port := config.Get("EMAIL_PORT")
	user := config.Get("EMAIL_HOST_USER")
	pass := config.Get("EMAIL_HOST_PASSWORD")
	from := config.Get("DEFAULT_FROM_EMAIL")

	auth := smtp.PlainAuth("", user, pass, host)

	subject := "Reset Password YourApp"
	body := fmt.Sprintf(`
<p>Hai,</p>
<p>Kamu meminta reset password. Klik link berikut untuk mengatur password baru:</p>
<p><a href="%s">%s</a></p>
<p>Link ini berlaku 1 jam.</p>
`, resetURL, resetURL)

	msg := []byte(
		fmt.Sprintf("From: %s\r\n", from) +
			fmt.Sprintf("To: %s\r\n", to) +
			fmt.Sprintf("Subject: %s\r\n", subject) +
			"Mime-Version: 1.0;\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
			fmt.Sprintf("%s\r\n", body),
	)

	addr := fmt.Sprintf("%s:%s", host, port)
	if err := smtp.SendMail(addr, auth, from, []string{to}, msg); err != nil {
		return err
	}

	log.Printf("Reset password email sent to %s with link: %s\n", to, resetURL)
	return nil
}
