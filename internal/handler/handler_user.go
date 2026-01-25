package handler

import (
	"context"
	"database/sql"
	"log"
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

// UserResponse struct untuk JSON response
type UserResponse struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// GetAllUsers menangani GET /users dengan pagination
func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	// Ambil pagination dari query params
	pagination := utils.GetPagination(c, 1, 10, 100)

	// Gunakan context dengan timeout
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// Hitung total user
	var total int
	if err := h.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		log.Printf("GetAllUsers: failed to count users: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to count users")
	}

	// Query data user dengan limit & offset
	rows, err := h.DB.QueryContext(ctx,
		"SELECT id, email, created_at FROM users ORDER BY id LIMIT $1 OFFSET $2",
		pagination.Limit, pagination.Offset,
	)
	if err != nil {
		log.Printf("GetAllUsers: failed to query users: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to query users")
	}
	defer rows.Close()

	users := []UserResponse{}

	for rows.Next() {
		var u UserResponse
		var createdAt time.Time

		if err := rows.Scan(&u.ID, &u.Email, &createdAt); err != nil {
			log.Printf("GetAllUsers: failed to scan user: %v", err)
			return utils.Error(c, fiber.StatusInternalServerError, "failed to scan user")
		}

		u.CreatedAt = createdAt.Format(time.RFC3339)
		users = append(users, u)
	}

	// Cek error iterasi
	if err := rows.Err(); err != nil {
		log.Printf("GetAllUsers: rows iteration error: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "error reading users")
	}

	// Buat response meta pagination dari utils.GetPaginatedResponse
	items, meta := utils.GetPaginatedResponse(users, total, pagination.Page, pagination.Limit)

	// Buat HATEOAS links opsional
	links := map[string]string{
		"next": "/users?page=2&limit=10", // bisa dihitung dinamis dari meta.TotalPages
		"prev": "",                        // kosong jika page=1
	}

	// Response pakai pro+ API response dengan message, meta, dan links
	return utils.SuccessMessage(c, "List users retrieved successfully", items, meta, links)
}
