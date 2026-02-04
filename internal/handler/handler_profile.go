// Package handler untuk endpoint profile
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

type ProfileHandler struct {
	DB *sql.DB
}

func NewProfileHandler(db *sql.DB) *ProfileHandler {
	return &ProfileHandler{DB: db}
}

// ProfileResponse untuk response
type ProfileResponse struct {
	ID           int    `json:"id"`
	Nama         string `json:"nama"`
	NamaBelakang string `json:"nama_belakang,omitempty"`
	TanggalLahir string `json:"tanggal_lahir"`
	Avatar       string `json:"avatar,omitempty"`
	IsVerified   bool   `json:"is_verified"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	UserID       int    `json:"user_id"`
}

// GetProfileByID untuk mendapatkan profile berdasarkan id user.
func (h *ProfileHandler) GetProfileByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid profile id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var p ProfileResponse
	var namaBelakang sql.NullString
	var avatar sql.NullString
	var createdAt, updatedAt sql.NullTime

	err = h.DB.QueryRowContext(ctx, `
		SELECT p.id, p.nama, p.nama_belakang, p.tanggal_lahir, p.avatar, p.is_verified,
		       p.created_at, p.updated_at, up.user_id
		FROM profiles p
		LEFT JOIN user_profiles up ON up.profile_id = p.id
		WHERE p.id = $1
	`, id).Scan(
		&p.ID, &p.Nama, &namaBelakang, &p.TanggalLahir, &avatar, &p.IsVerified,
		&createdAt, &updatedAt, &p.UserID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return utils.Error(c, fiber.StatusNotFound, "profile not found")
		}
		log.Printf("GetProfileByID: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to get profile")
	}

	if namaBelakang.Valid {
		p.NamaBelakang = namaBelakang.String
	}
	if avatar.Valid {
		p.Avatar = avatar.String
	}
	p.CreatedAt = createdAt.Time.Format(time.RFC3339)
	p.UpdatedAt = updatedAt.Time.Format(time.RFC3339)

	return utils.SuccessMessage(c, "Profile retrieved successfully", p, nil, nil)
}

// GetMyProfile digunakan untuk mendapatkan profile yang login.
func (h *ProfileHandler) GetMyProfile(c *fiber.Ctx) error {
	// ambil user_id dari token (middleware nanti set ke locals)
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.Error(c, fiber.StatusUnauthorized, "unauthorized")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var p ProfileResponse
	var namaBelakang sql.NullString
	var avatar sql.NullString
	var createdAt, updatedAt sql.NullTime

	err := h.DB.QueryRowContext(ctx, `
        SELECT p.id, p.nama, p.nama_belakang, p.tanggal_lahir, p.avatar, p.is_verified,
               p.created_at, p.updated_at, up.user_id
        FROM profiles p
        JOIN user_profiles up ON up.profile_id = p.id
        WHERE up.user_id = $1
    `, userID).Scan(
		&p.ID, &p.Nama, &namaBelakang, &p.TanggalLahir, &avatar, &p.IsVerified,
		&createdAt, &updatedAt, &p.UserID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return utils.Error(c, fiber.StatusNotFound, "profile not found")
		}
		log.Printf("GetMyProfile: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to get profile")
	}

	if namaBelakang.Valid {
		p.NamaBelakang = namaBelakang.String
	}
	if avatar.Valid {
		p.Avatar = avatar.String
	}
	p.CreatedAt = createdAt.Time.Format(time.RFC3339)
	p.UpdatedAt = updatedAt.Time.Format(time.RFC3339)

	return utils.SuccessMessage(c, "Profile retrieved successfully", p, nil, nil)
}

// GetAllProfiles untuk mendapatkan semua profile.
func (h *ProfileHandler) GetAllProfiles(c *fiber.Ctx) error {
	pagination := utils.GetPagination(c, 1, 10, 100)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var total int
	if err := h.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM profiles").Scan(&total); err != nil {
		log.Printf("GetAllProfiles: failed to count profiles: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to count profiles")
	}

	rows, err := h.DB.QueryContext(ctx, `
		SELECT p.id, p.nama, p.nama_belakang, p.tanggal_lahir, p.avatar, p.is_verified,
		       p.created_at, p.updated_at, COALESCE(up.user_id, 0) as user_id
		FROM profiles p
		LEFT JOIN user_profiles up ON up.profile_id = p.id
		ORDER BY p.id
		LIMIT $1 OFFSET $2
	`, pagination.Limit, pagination.Offset)

	if err != nil {
		log.Printf("GetAllProfiles: failed to query profiles: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to query profiles")
	}
	defer rows.Close()

	profiles := []ProfileResponse{}

	for rows.Next() {
		var p ProfileResponse
		var namaBelakang sql.NullString
		var avatar sql.NullString
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&p.ID, &p.Nama, &namaBelakang, &p.TanggalLahir, &avatar, &p.IsVerified, &createdAt, &updatedAt, &p.UserID); err != nil {
			log.Printf("GetAllProfiles: failed to scan profile: %v", err)
			return utils.Error(c, fiber.StatusInternalServerError, "failed to scan profile")
		}

		if namaBelakang.Valid {
			p.NamaBelakang = namaBelakang.String
		}
		if avatar.Valid {
			p.Avatar = avatar.String
		}

		p.CreatedAt = createdAt.Time.Format(time.RFC3339)
		p.UpdatedAt = updatedAt.Time.Format(time.RFC3339)

		profiles = append(profiles, p)
	}

	if err := rows.Err(); err != nil {
		log.Printf("GetAllProfiles: rows iteration error: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "error reading profiles")
	}

	items, meta := utils.GetPaginatedResponse(profiles, total, pagination.Page, pagination.Limit)

	links := map[string]string{"next": "", "prev": ""}
	if pagination.Page < meta.TotalPages {
		links["next"] = fmt.Sprintf("/profiles?page=%d&limit=%d", pagination.Page+1, pagination.Limit)
	}
	if pagination.Page > 1 {
		links["prev"] = fmt.Sprintf("/profiles?page=%d&limit=%d", pagination.Page-1, pagination.Limit)
	}

	return utils.SuccessMessage(c, "Profiles retrieved successfully", items, meta, links)
}

// GetProfileByUserID untuk mendapatkan profile user berdasarkan id user.
func (h *ProfileHandler) GetProfileByUserID(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	var p ProfileResponse
	var namaBelakang sql.NullString
	var avatar sql.NullString
	var createdAt, updatedAt sql.NullTime

	err = h.DB.QueryRowContext(ctx, `
		SELECT p.id, p.nama, p.nama_belakang, p.tanggal_lahir, p.avatar, p.is_verified,
		       p.created_at, p.updated_at, up.user_id
		FROM profiles p
		JOIN user_profiles up ON up.profile_id = p.id
		WHERE up.user_id = $1
	`, userID).Scan(
		&p.ID, &p.Nama, &namaBelakang, &p.TanggalLahir, &avatar, &p.IsVerified,
		&createdAt, &updatedAt, &p.UserID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return utils.Error(c, fiber.StatusNotFound, "profile not found")
		}
		log.Printf("GetProfileByUserID: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to get profile")
	}

	if namaBelakang.Valid {
		p.NamaBelakang = namaBelakang.String
	}
	if avatar.Valid {
		p.Avatar = avatar.String
	}
	p.CreatedAt = createdAt.Time.Format(time.RFC3339)
	p.UpdatedAt = updatedAt.Time.Format(time.RFC3339)

	return utils.SuccessMessage(c, "Profile retrieved successfully", p, nil, nil)
}

// CreateProfileRequest adalah model untuk membuat profile.
type CreateProfileRequest struct {
	Nama         string  `json:"nama"`
	NamaBelakang *string `json:"nama_belakang"`
	TanggalLahir string  `json:"tanggal_lahir"` // format YYYY-MM-DD
	IsVerified   *bool   `json:"is_verified"`
}

func (h *ProfileHandler) CreateProfileByUserID(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	// cek apakah sudah ada profil, karena ini harus unik
	var existingProfileID int
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	err = h.DB.QueryRowContext(ctx, `
		SELECT profile_id FROM user_profiles WHERE user_id = $1
	`, userID).Scan(&existingProfileID)

	if err == nil {
		// profil sudah ada
		return utils.Error(c, fiber.StatusForbidden, "Profile already exists, use PUT to update")
	} else if err != sql.ErrNoRows {
		// error lain
		log.Printf("check existing profile: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "internal error")
	}

	// parse form-data (bukan JSON)
	nama := c.FormValue("nama")
	namaBelakang := c.FormValue("nama_belakang")
	tanggalLahir := c.FormValue("tanggal_lahir")
	isVerifiedStr := c.FormValue("is_verified")

	if nama == "" {
		return utils.Error(c, fiber.StatusBadRequest, "nama is required")
	}
	if tanggalLahir == "" {
		return utils.Error(c, fiber.StatusBadRequest, "tanggal_lahir is required")
	}
	if _, err := time.Parse("2006-01-02", tanggalLahir); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "tanggal_lahir must be YYYY-MM-DD")
	}

	var isVerified *bool
	if isVerifiedStr != "" {
		v := isVerifiedStr == "true" || isVerifiedStr == "1"
		isVerified = &v
	}

	var namaBelakangPtr *string
	if namaBelakang != "" {
		namaBelakangPtr = &namaBelakang
	}

	// handle avatar upload (optional)
	file, err := c.FormFile("avatar")
	avatarPath := ""
	if err == nil {
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
		avatarPath = fmt.Sprintf("/media/avatars/%s", filename)

		if err := c.SaveFile(file, fmt.Sprintf("./media/avatars/%s", filename)); err != nil {
			log.Printf("CreateProfileByUserID save file: %v", err)
			return utils.Error(c, fiber.StatusInternalServerError, "failed to upload avatar")
		}
	}

	var profileID int
	err = h.DB.QueryRowContext(ctx, `
		INSERT INTO profiles (nama, nama_belakang, tanggal_lahir, avatar, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`, nama, namaBelakangPtr, tanggalLahir, avatarPath, isVerified).Scan(&profileID)

	if err != nil {
		log.Printf("CreateProfileByUserID: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to create profile")
	}

	_, err = h.DB.ExecContext(ctx, `
		INSERT INTO user_profiles (user_id, profile_id, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
	`, userID, profileID)

	if err != nil {
		log.Printf("CreateProfileByUserID (relation): %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to link profile to user")
	}

	return utils.SuccessMessage(c, "Profile created successfully", map[string]int{"id": profileID}, nil, nil)
}

// UpdateProfileRequest model profile update yang diperlukan.
type UpdateProfileRequest struct {
	Nama         *string `json:"nama"`
	NamaBelakang *string `json:"nama_belakang"`
	TanggalLahir *string `json:"tanggal_lahir"`
	IsVerified   *bool   `json:"is_verified"`
}

func (h *ProfileHandler) UpdateProfileByUserID(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	// parse form-data (bukan JSON)
	nama := c.FormValue("nama")
	namaBelakang := c.FormValue("nama_belakang")
	tanggalLahir := c.FormValue("tanggal_lahir")
	isVerifiedStr := c.FormValue("is_verified")

	var isVerified *bool
	if isVerifiedStr != "" {
		v := isVerifiedStr == "true" || isVerifiedStr == "1"
		isVerified = &v
	}

	var namaBelakangPtr *string
	if namaBelakang != "" {
		namaBelakangPtr = &namaBelakang
	}

	// handle avatar upload (optional)
	file, err := c.FormFile("avatar")
	avatarPath := ""
	if err == nil {
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
		avatarPath = fmt.Sprintf("/media/avatars/%s", filename)

		if err := c.SaveFile(file, fmt.Sprintf("./media/avatars/%s", filename)); err != nil {
			log.Printf("UpdateProfileByUserID save file: %v", err)
			return utils.Error(c, fiber.StatusInternalServerError, "failed to upload avatar")
		}
	}

	if nama == "" && namaBelakang == "" && tanggalLahir == "" && avatarPath == "" && isVerified == nil {
		return utils.Error(c, fiber.StatusBadRequest, "nothing to update")
	}

	if tanggalLahir != "" {
		if _, err := time.Parse("2006-01-02", tanggalLahir); err != nil {
			return utils.Error(c, fiber.StatusBadRequest, "tanggal_lahir must be YYYY-MM-DD")
		}
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	query := "UPDATE profiles SET "
	args := []interface{}{}
	i := 1

	if nama != "" {
		query += fmt.Sprintf("nama = $%d, ", i)
		args = append(args, nama)
		i++
	}
	if namaBelakangPtr != nil {
		query += fmt.Sprintf("nama_belakang = $%d, ", i)
		args = append(args, *namaBelakangPtr)
		i++
	}
	if tanggalLahir != "" {
		query += fmt.Sprintf("tanggal_lahir = $%d, ", i)
		args = append(args, tanggalLahir)
		i++
	}
	if avatarPath != "" {
		query += fmt.Sprintf("avatar = $%d, ", i)
		args = append(args, avatarPath)
		i++
	}
	if isVerified != nil {
		query += fmt.Sprintf("is_verified = $%d, ", i)
		args = append(args, *isVerified)
		i++
	}

	query += fmt.Sprintf("updated_at = NOW() WHERE id = (SELECT profile_id FROM user_profiles WHERE user_id = $%d)", i)
	args = append(args, userID)

	res, err := h.DB.ExecContext(ctx, query, args...)
	if err != nil {
		log.Printf("UpdateProfileByUserID: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to update profile")
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return utils.Error(c, fiber.StatusNotFound, "profile not found")
	}

	return utils.SuccessMessage(c, "Profile updated successfully", nil, nil, nil)
}

// DeleteProfileByUserID untuk menghapus profile berdasarkan user id.
func (h *ProfileHandler) DeleteProfileByUserID(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid user id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	_, err = h.DB.ExecContext(ctx, "DELETE FROM user_profiles WHERE user_id = $1", userID)
	if err != nil {
		log.Printf("DeleteProfileByUserID relation: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to delete profile relation")
	}

	// delete profile
	res, err := h.DB.ExecContext(ctx, `
		DELETE FROM profiles 
		WHERE id = (SELECT profile_id FROM user_profiles WHERE user_id = $1)
	`, userID)

	if err != nil {
		log.Printf("DeleteProfileByUserID: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to delete profile")
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return utils.Error(c, fiber.StatusNotFound, "profile not found")
	}

	return utils.SuccessMessage(c, "Profile deleted successfully", nil, nil, nil)
}

// GetProfileByAdmin hanya bisa diakses admin, melihat profile user mana pun
func (h *ProfileHandler) GetProfileByAdmin(c *fiber.Ctx) error {
	// ambil param id profile yang ingin dilihat
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid profile id")
	}

	// ambil user_id dari context (set oleh middleware Auth)
	userIDLocal := c.Locals("user_id")
	var userID int
	switch v := userIDLocal.(type) {
	case int:
		userID = v
	case int64:
		userID = int(v)
	case float64:
		userID = int(v)
	case string:
		userID, err = strconv.Atoi(v)
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "invalid user_id")
		}
	default:
		return utils.Error(c, fiber.StatusInternalServerError, "invalid user_id type")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// cek apakah user adalah admin dari tabel roles + user_roles
	var isAdmin bool
	err = h.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM roles r
			JOIN user_roles ur ON ur.role_id = r.id
			WHERE ur.user_id = $1 AND r.name = 'admin'
		)
	`, userID).Scan(&isAdmin)
	if err != nil {
		log.Printf("Check admin role: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to check role")
	}
	if !isAdmin {
		return utils.Error(c, fiber.StatusForbidden, "only admin can access this endpoint")
	}

	// ambil profile user berdasarkan id
	var p ProfileResponse
	var namaBelakang sql.NullString
	var avatar sql.NullString
	var createdAt, updatedAt sql.NullTime

	err = h.DB.QueryRowContext(ctx, `
		SELECT p.id, p.nama, p.nama_belakang, p.tanggal_lahir, p.avatar, p.is_verified,
		       p.created_at, p.updated_at, up.user_id
		FROM profiles p
		LEFT JOIN user_profiles up ON up.profile_id = p.id
		WHERE p.id = $1
	`, id).Scan(
		&p.ID, &p.Nama, &namaBelakang, &p.TanggalLahir, &avatar, &p.IsVerified,
		&createdAt, &updatedAt, &p.UserID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.Error(c, fiber.StatusNotFound, "profile not found")
		}
		log.Printf("GetProfileByAdmin: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to get profile")
	}

	if namaBelakang.Valid {
		p.NamaBelakang = namaBelakang.String
	}
	if avatar.Valid {
		p.Avatar = avatar.String
	}
	p.CreatedAt = createdAt.Time.Format(time.RFC3339)
	p.UpdatedAt = updatedAt.Time.Format(time.RFC3339)

	return utils.SuccessMessage(c, "Profile retrieved successfully", p, nil, nil)
}
