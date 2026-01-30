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

type UserRoleHandler struct {
	DB *sql.DB
}

func NewUserRoleHandler(db *sql.DB) *UserRoleHandler {
	return &UserRoleHandler{DB: db}
}

type UserRoleResponse struct {
	UserID int      `json:"user_id"`
	Roles  []string `json:"roles"`
}

type AssignRoleRequest struct {
	RoleID int `json:"role_id"`
}

// ============================
// GET /users/:id/roles
// ============================
func (h *UserRoleHandler) GetUserRoles(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	pagination := utils.GetPagination(c, 1, 10, 100)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// total roles
	var total int
	if err := h.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_roles WHERE user_id = $1`, userID).Scan(&total); err != nil {
		log.Printf("GetUserRoles: failed to count roles: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to count roles")
	}

	rows, err := h.DB.QueryContext(ctx,
		`SELECT r.id, r.name
		 FROM roles r
		 JOIN user_roles ur ON ur.role_id = r.id
		 WHERE ur.user_id = $1
		 ORDER BY r.id
		 LIMIT $2 OFFSET $3`,
		userID, pagination.Limit, pagination.Offset,
	)
	if err != nil {
		log.Printf("GetUserRoles: failed to query roles: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to query roles")
	}
	defer rows.Close()

	roles := []RoleResponse{}
	for rows.Next() {
		var r RoleResponse
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			log.Printf("GetUserRoles: failed to scan role: %v", err)
			return utils.Error(c, fiber.StatusInternalServerError, "failed to scan roles")
		}
		roles = append(roles, r)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetUserRoles: rows iteration error: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "error reading roles")
	}

	items, meta := utils.GetPaginatedResponse(roles, total, pagination.Page, pagination.Limit)

	links := map[string]string{"next": "", "prev": ""}
	if pagination.Page < meta.TotalPages {
		links["next"] = fmt.Sprintf("/users/%d/roles?page=%d&limit=%d", userID, pagination.Page+1, pagination.Limit)
	}
	if pagination.Page > 1 {
		links["prev"] = fmt.Sprintf("/users/%d/roles?page=%d&limit=%d", userID, pagination.Page-1, pagination.Limit)
	}

	return utils.SuccessMessage(c, "user roles retrieved successfully", items, meta, links)
}

// ============================
// PUT /users/:id/role
// replace role user (user wajib punya role)
// ============================
func (h *UserRoleHandler) UpdateUserRole(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	var req AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("UpdateUserRole: failed to begin tx: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to update role")
	}

	// delete old role(s)
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM user_roles WHERE user_id = $1`, userID,
	); err != nil {
		tx.Rollback()
		log.Printf("UpdateUserRole: failed to delete old role: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to update role")
	}

	// insert new role
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
		userID, req.RoleID,
	); err != nil {
		tx.Rollback()
		log.Printf("UpdateUserRole: failed to insert new role: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to update role")
	}

	if err := tx.Commit(); err != nil {
		log.Printf("UpdateUserRole: commit failed: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to update role")
	}

	return utils.SuccessMessage(c, "user role updated successfully", nil, nil)
}

// ============================
// POST /users/:id/roles
// assign role baru (optional)
// ============================
func (h *UserRoleHandler) AssignRole(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	var req AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	_, err = h.DB.ExecContext(ctx,
		`INSERT INTO user_roles (user_id, role_id)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id, role_id) DO NOTHING`,
		userID, req.RoleID,
	)
	if err != nil {
		log.Printf("AssignRole: failed to assign role: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to assign role")
	}

	return utils.SuccessMessage(c, "role assigned successfully", nil, nil)
}

// ============================
// DELETE /users/:id/roles/:roleId
// optional (kalau kamu mau remove role tertentu)
// ============================
func (h *UserRoleHandler) RemoveRole(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	roleID, err := strconv.Atoi(c.Params("roleId"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid role id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	result, err := h.DB.ExecContext(ctx,
		`DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`,
		userID, roleID,
	)
	if err != nil {
		log.Printf("RemoveRole: failed to remove role: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to remove role")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return utils.Error(c, fiber.StatusNotFound, "role not found for this user")
	}

	return utils.SuccessMessage(c, "role removed successfully", nil, nil)
}
