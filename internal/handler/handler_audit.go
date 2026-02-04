// Package handler berisi handler untuk berbagai resource
package handler

import (
	"database/sql"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AuditHandler struct {
	DB *sql.DB
}

func NewAuditHandler(db *sql.DB) *AuditHandler {
	return &AuditHandler{DB: db}
}

// AuditLogResponse merepresentasikan satu entri audit log
type AuditLogResponse struct {
	ID        int       `json:"id"`
	UserID    any       `json:"user_id"`
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	Status    int       `json:"status"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

// GetAuditLogs GET /audit-logs
func (h *AuditHandler) GetAuditLogs(c *fiber.Ctx) error {
	rows, err := h.DB.Query(`
		SELECT id, user_id, method, url, status, ip, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		log.Printf("GetAuditLogs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to query audit logs",
		})
	}
	defer rows.Close()

	var logs []AuditLogResponse
	for rows.Next() {
		var a AuditLogResponse
		if err := rows.Scan(&a.ID, &a.UserID, &a.Method, &a.URL, &a.Status, &a.IP, &a.CreatedAt); err != nil {
			log.Printf("GetAuditLogs scan: %v", err)
			continue
		}
		logs = append(logs, a)
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"data":   logs,
	})
}
