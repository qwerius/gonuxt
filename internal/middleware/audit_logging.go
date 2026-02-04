// Package middleware berisi middleware untuk logging dan auditing
package middleware

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// AuditConfig berisi konfigurasi audit
type AuditConfig struct {
	DB *sql.DB
}

// AuditLoggingMiddleware membuat middleware audit logging
func AuditLoggingMiddleware(cfg *AuditConfig) fiber.Handler {
	// daftar route yang tidak dicatat
	skipPaths := map[string]bool{
		"/api/v1/auth/login":           true,
		"/api/v1/auth/register":        true,
		"/api/v1/auth/refresh":         true,
		"/api/v1/auth/forgot-password": true,
		"/api/v1/auth/reset-password":  true,
		"/oauth/google/login":          true,
		"/oauth/google/callback":       true,
	}

	return func(c *fiber.Ctx) error {
		// cek apakah route ini dikecualikan
		if skipPaths[c.Path()] {
			return c.Next()
		}

		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		// ambil userID dari context, diasumsikan sudah ada
		userID := c.Locals("user_id")

		// log ke console
		log.Printf("[AUDIT] user=%v ip=%s method=%s url=%s status=%d duration=%s",
			userID, c.IP(), c.Method(), c.OriginalURL(), c.Response().StatusCode(), duration,
		)

		// simpan ke DB jika ada koneksi
		if cfg.DB != nil {
			go saveAuditToDB(cfg.DB, userID, c.Method(), c.OriginalURL(), c.Response().StatusCode(), c.IP())
		}

		return err
	}
}

// simpan log ke DB
func saveAuditToDB(db *sql.DB, userID any, method, url string, status int, ip string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx,
		`INSERT INTO audit_logs (user_id, method, url, status, ip)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, method, url, status, ip,
	)
	if err != nil {
		log.Printf("saveAuditToDB error: %v", err)
	}
}
