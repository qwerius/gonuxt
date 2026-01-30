package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// SendEmailSMTP mengirim email pakai SMTP dari env
func SendEmailSMTP(to, subject, body string) error {
	host := os.Getenv("EMAIL_HOST")
	port := os.Getenv("EMAIL_PORT")
	user := os.Getenv("EMAIL_HOST_USER")
	pass := os.Getenv("EMAIL_HOST_PASSWORD")
	from := os.Getenv("DEFAULT_FROM_EMAIL")

	auth := smtp.PlainAuth("", user, pass, host)

	msg := []byte(
		fmt.Sprintf("From: %s\r\n", from) +
			fmt.Sprintf("To: %s\r\n", to) +
			fmt.Sprintf("Subject: %s\r\n", subject) +
			"Mime-Version: 1.0;\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
			fmt.Sprintf("%s\r\n", body),
	)

	addr := fmt.Sprintf("%s:%s", host, port)
	err := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		log.Println("Email send failed:", err)
		return err
	}

	log.Println("Email sent to", to)
	return nil
}
