package handler

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/utils"
)

type UserHandler struct {
	DB *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{DB: db}
}

// GetAllUsers menangani GET /users dengan pagination
func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	pagination := utils.GetPagination(c, 1, 10, 100) // default page=1, limit=10, maxLimit=100

	var total int
	if err := h.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to count users")
	}

	rows, err := h.DB.Query(
		"SELECT id, email, created_at FROM users ORDER BY id LIMIT $1 OFFSET $2",
		pagination.Limit, pagination.Offset,
	)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to query users")
	}
	defer rows.Close()

	users := []map[string]interface{}{}

	for rows.Next() {
		var id int
		var email string
		var createdAt time.Time
		if err := rows.Scan(&id, &email, &createdAt); err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "failed to scan user")
		}
		users = append(users, map[string]interface{}{
			"id":         id,
			"email":      email,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}

	// Response pakai response.go + paginasi.go
	return utils.Success(c, utils.GetPaginatedResponse(users, total, pagination.Page, pagination.Limit))
}
