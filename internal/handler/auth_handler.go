package handler

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler struct
type AuthHandler struct {
	DB *sql.DB
}

// NewAuthHandler constructor
func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

// RegisterRequest payload
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshTokenRequest payload untuk body JSON
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Register user baru
func (h *AuthHandler) Register(c *fiber.Ctx) error {
    // Cek body kosong dulu
    if len(c.Body()) == 0 {
        return utils.Error(c, fiber.StatusBadRequest, "Request body is required")
    }

    var body RegisterRequest
    if err := c.BodyParser(&body); err != nil {
        return utils.Error(c, fiber.StatusBadRequest, "Invalid request body: must be valid JSON")
    }

    // Validasi field
    if body.Email == "" {
        return utils.Error(c, fiber.StatusBadRequest, "Email is required")
    }
    if body.Password == "" {
        return utils.Error(c, fiber.StatusBadRequest, "Password is required")
    }

    // Cek apakah email sudah ada
    var exists bool
    if err := h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", body.Email).Scan(&exists); err != nil {
        log.Printf("Register: failed to check email existence: %v", err)
        return utils.Error(c, fiber.StatusInternalServerError, "Database error")
    }
    if exists {
        return utils.Error(c, fiber.StatusConflict, "Email already registered")
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
    if err != nil {
        log.Printf("Register: failed to hash password: %v", err)
        return utils.Error(c, fiber.StatusInternalServerError, "Failed to create user")
    }

    // Simpan user
    var id int
    err = h.DB.QueryRow(
        "INSERT INTO users (email, password, created_at) VALUES ($1, $2, $3) RETURNING id",
        body.Email, string(hashedPassword), time.Now(),
    ).Scan(&id)
    if err != nil {
        log.Printf("Register: failed to insert user: %v", err)
        return utils.Error(c, fiber.StatusInternalServerError, "Failed to create user")
    }

    // Response success
    return utils.SuccessMessage(c, "User registered successfully", map[string]interface{}{
        "id":    id,
        "email": body.Email,
    }, nil)
}

// Login user
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var body LoginRequest
	if err := c.BodyParser(&body); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Request body is required")
	}

	if body.Email == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Email is required")
	}
	if body.Password == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Password is required")
	}

	// Ambil user dari DB
	var id int
	var hashedPassword string
	err := h.DB.QueryRow("SELECT id, password FROM users WHERE email=$1", body.Email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.Error(c, fiber.StatusUnauthorized, "Invalid email or password")
		}
		log.Printf("Login: failed to query user: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "Database error")
	}

	// Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(body.Password)); err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "Invalid email or password")
	}

	// Buat access token
	accessToken, err := utils.CreateAccessToken(id)
	if err != nil {
		log.Printf("Login: failed to create access token: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to create access token")
	}

	// Buat refresh token
	refreshToken, err := utils.CreateRefreshToken(id)
	if err != nil {
		log.Printf("Login: failed to create refresh token: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to create refresh token")
	}

	// Response success dengan pro format + meta optional
	return utils.SuccessMessage(c, "Login successful", map[string]interface{}{
		"access_token":       accessToken,
		"refresh_token":      refreshToken,
		"token_type":         "Bearer",
		"expires_in":         3600,      // 1 jam
		"refresh_expires_in": 604800,    // 7 hari
	}, nil)
}

// RefreshToken handler
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var body RefreshTokenRequest
	if err := c.BodyParser(&body); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if body.RefreshToken == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Refresh token is required")
	}

	userID, err := utils.ValidateRefreshToken(body.RefreshToken)
	if err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "Invalid refresh token")
	}

	accessToken, err := utils.CreateAccessToken(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to create access token")
	}

	newRefreshToken, err := utils.CreateRefreshToken(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to create refresh token")
	}

	return utils.SuccessMessage(c, "Token refreshed successfully", map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
	}, nil)
}
