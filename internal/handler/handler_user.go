package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/utils"
	"golang.org/x/crypto/bcrypt"
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
	UpdatedAt string `json:"updated_at"`
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

	// Query data user dengan COALESCE untuk updated_at
	rows, err := h.DB.QueryContext(ctx,
		`SELECT id, email, created_at, COALESCE(updated_at, created_at) AS updated_at
		 FROM users
		 ORDER BY id
		 LIMIT $1 OFFSET $2`,
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
		var createdAt, updatedAt sql.NullTime

		// Scan semua kolom sekaligus, pakai NullTime untuk aman
		if err := rows.Scan(&u.ID, &u.Email, &createdAt, &updatedAt); err != nil {
			log.Printf("GetAllUsers: failed to scan user: %v", err)
			return utils.Error(c, fiber.StatusInternalServerError, "failed to scan user")
		}

		// Format created_at dan updated_at ke RFC3339
		if createdAt.Valid {
			u.CreatedAt = createdAt.Time.Format(time.RFC3339)
		} else {
			u.CreatedAt = ""
		}

		// updatedAt dijamin valid karena COALESCE di query
		if updatedAt.Valid {
			u.UpdatedAt = updatedAt.Time.Format(time.RFC3339)
		} else {
			u.UpdatedAt = ""
		}

		users = append(users, u)
	}

	// Cek error iterasi
	if err := rows.Err(); err != nil {
		log.Printf("GetAllUsers: rows iteration error: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "error reading users")
	}

	// Buat response meta pagination dari utils.GetPaginatedResponse
	items, meta := utils.GetPaginatedResponse(users, total, pagination.Page, pagination.Limit)

	// Buat HATEOAS links dinamis
	links := map[string]string{
		"next": "",
		"prev": "",
	}
	if pagination.Page < meta.TotalPages {
		links["next"] = fmt.Sprintf("/users?page=%d&limit=%d", pagination.Page+1, pagination.Limit)
	}
	if pagination.Page > 1 {
		links["prev"] = fmt.Sprintf("/users?page=%d&limit=%d", pagination.Page-1, pagination.Limit)
	}

	// Response pakai pro+ API response dengan message, meta, dan links
	return utils.SuccessMessage(c, "List users retrieved successfully", items, meta, links)
}

// GetUserByID untuk mendapatkan user berdasarkan id usernya
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var user UserResponse
	var createdAt, updatedAt sql.NullTime

	err = h.DB.QueryRowContext(ctx,
		`SELECT id, email, created_at, COALESCE(updated_at, created_at)
		 FROM users
		 WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &createdAt, &updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return utils.Error(c, fiber.StatusNotFound, "user not found")
		}
		log.Printf("GetUserByID: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to get user")
	}

	user.CreatedAt = createdAt.Time.Format(time.RFC3339)
	user.UpdatedAt = updatedAt.Time.Format(time.RFC3339)

	return utils.SuccessMessage(c, "User retrieved successfully", user, nil, nil)
}

// CreateUserRequest struct membuat user.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"` // plain text from client
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var body CreateUserRequest
	if err := c.BodyParser(&body); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	// Hash password sebelum simpan
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to hash password")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var id int
	err = h.DB.QueryRowContext(ctx,
		`INSERT INTO users (email, password, created_at)
		 VALUES ($1, $2, NOW())
		 RETURNING id`,
		body.Email, string(hashedPassword),
	).Scan(&id)

	if err != nil {
		log.Printf("CreateUser: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to create user")
	}

	return utils.SuccessMessage(c, "User created successfully", map[string]int{"id": id}, nil, nil)
}

// UpdateUserRequest untuk mengubah user berdasarkan id nya.
type UpdateUserRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"` // plain text from client
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	var body UpdateUserRequest
	if err := c.BodyParser(&body); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	if body.Email == nil && body.Password == nil {
		return utils.Error(c, fiber.StatusBadRequest, "nothing to update")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	query := "UPDATE users SET "
	args := []interface{}{}
	i := 1

	if body.Email != nil {
		query += fmt.Sprintf("email = $%d, ", i)
		args = append(args, *body.Email)
		i++
	}

	if body.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*body.Password), bcrypt.DefaultCost)
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "failed to hash password")
		}
		query += fmt.Sprintf("password = $%d, ", i)
		args = append(args, string(hashedPassword))
		i++
	}

	query += fmt.Sprintf("updated_at = NOW() WHERE id = $%d", i)
	args = append(args, id)

	res, err := h.DB.ExecContext(ctx, query, args...)
	if err != nil {
		log.Printf("UpdateUser: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to update user")
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return utils.Error(c, fiber.StatusNotFound, "user not found")
	}

	return utils.SuccessMessage(c, "User updated successfully", nil, nil, nil)
}

// DeleteUser untuk menghapus user berdasarkan id nya.
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	res, err := h.DB.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		log.Printf("DeleteUser: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to delete user")
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return utils.Error(c, fiber.StatusNotFound, "user not found")
	}

	return utils.SuccessMessage(c, "User deleted successfully", nil, nil, nil)
}
