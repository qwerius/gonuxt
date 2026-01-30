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
)

type RoleHandler struct {
	DB *sql.DB
}

func NewRoleHandler(db *sql.DB) *RoleHandler {
	return &RoleHandler{DB: db}
}

type RoleResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateRoleRequest struct {
	Name string `json:"name"`
}

// GetAllRoles GET /roles (dengan pagination)
func (h *RoleHandler) GetAllRoles(c *fiber.Ctx) error {
	pagination := utils.GetPagination(c, 1, 10, 100)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// hitung total role
	var total int
	if err := h.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM roles").Scan(&total); err != nil {
		log.Printf("GetAllRoles: failed to count roles: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to count roles")
	}

	// query role dengan limit offset
	rows, err := h.DB.QueryContext(ctx,
		`SELECT id, name
		 FROM roles
		 ORDER BY id
		 LIMIT $1 OFFSET $2`,
		pagination.Limit, pagination.Offset,
	)
	if err != nil {
		log.Printf("GetAllRoles: failed to query roles: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to query roles")
	}
	defer rows.Close()

	roles := []RoleResponse{}
	for rows.Next() {
		var r RoleResponse
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			log.Printf("GetAllRoles: failed to scan role: %v", err)
			return utils.Error(c, fiber.StatusInternalServerError, "failed to scan roles")
		}
		roles = append(roles, r)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetAllRoles: rows iteration error: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "error reading roles")
	}

	// response meta pagination
	items, meta := utils.GetPaginatedResponse(roles, total, pagination.Page, pagination.Limit)

	// links HATEOAS
	links := map[string]string{
		"next": "",
		"prev": "",
	}
	if pagination.Page < meta.TotalPages {
		links["next"] = fmt.Sprintf("/roles?page=%d&limit=%d", pagination.Page+1, pagination.Limit)
	}
	if pagination.Page > 1 {
		links["prev"] = fmt.Sprintf("/roles?page=%d&limit=%d", pagination.Page-1, pagination.Limit)
	}

	return utils.SuccessMessage(c, "roles retrieved successfully", items, meta, links)
}

// GetRoleByID GET /roles/:id
func (h *RoleHandler) GetRoleByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid role id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var r RoleResponse
	err = h.DB.QueryRowContext(ctx, `
		SELECT id, name
		FROM roles
		WHERE id = $1
	`, id).Scan(&r.ID, &r.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			return utils.Error(c, fiber.StatusNotFound, "role not found")
		}
		log.Printf("GetRoleByID: failed to query role: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to query role")
	}

	return utils.SuccessMessage(c, "role retrieved successfully", r, nil)
}

// DeleteRole DELETE /roles/:id
func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid role id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	result, err := h.DB.ExecContext(ctx, `
		DELETE FROM roles
		WHERE id = $1
	`, id)
	if err != nil {
		log.Printf("DeleteRole: failed to delete role: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to delete role")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("DeleteRole: failed to get rows affected: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to delete role")
	}
	if rowsAffected == 0 {
		return utils.Error(c, fiber.StatusNotFound, "role not found")
	}

	return utils.SuccessMessage(c, "role deleted successfully", nil, nil)
}

// CreateRole POST /roles
func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	var req CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	// validasi sederhana
	if req.Name == "" {
		return utils.Error(c, fiber.StatusBadRequest, "role name is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var id int
	err := h.DB.QueryRowContext(ctx, `
		INSERT INTO roles (name)
		VALUES ($1)
		RETURNING id
	`, req.Name).Scan(&id)

	if err != nil {
		log.Printf("CreateRole: failed to insert role: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to create role")
	}

	role := RoleResponse{
		ID:   id,
		Name: req.Name,
	}

	return utils.SuccessMessage(c, "role created successfully", role, nil)
}
