package utils

import (
	"github.com/gofiber/fiber/v2"
	"time"
	"github.com/google/uuid"
)

// APIResponse standar response untuk API
type APIResponse struct {
	Status    string      `json:"status"`             // "ok" atau "error"
	Code      int         `json:"code"`               // HTTP status code
	Message   string      `json:"message,omitempty"`  // pesan opsional
	Data      interface{} `json:"data,omitempty"`     // payload
	Meta      interface{} `json:"meta,omitempty"`     // pagination atau info tambahan
	Links     interface{} `json:"links,omitempty"`    // HATEOAS links
	Timestamp string      `json:"timestamp"`          // waktu server
	RequestID string      `json:"request_id"`         // ID unik request
}

// generateRequestID membuat UUID untuk request
func generateRequestID() string {
	return uuid.New().String()
}

// SuccessMessage mengembalikan response sukses dengan message, data, meta, links opsional
func SuccessMessage(c *fiber.Ctx, msg string, data interface{}, meta interface{}, links ...interface{}) error {
	resp := APIResponse{
		Status:    "ok",
		Code:      fiber.StatusOK,
		Message:   msg,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: generateRequestID(),
	}

	if len(links) > 0 {
		resp.Links = links[0]
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// Error mengembalikan response error dengan status code, message, dan optional data
func Error(c *fiber.Ctx, status int, msg string, data ...interface{}) error {
	resp := APIResponse{
		Status:    "error",
		Code:      status,
		Message:   msg,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: generateRequestID(),
	}

	if len(data) > 0 {
		resp.Data = data[0]
	}

	return c.Status(status).JSON(resp)
}
