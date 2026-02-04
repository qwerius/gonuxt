// Package handler auth, lengkap dengan captcha di login, register, forgot-password
package handler

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/middleware"
	"github.com/qwerius/gonuxt/internal/store"
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
	Email         string `json:"email"`
	Password      string `json:"password"`
	CaptchaID     string `json:"captcha_id"`
	CaptchaAnswer string `json:"captcha_answer"`
}

// LoginRequest payload
type LoginRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	CaptchaID     string `json:"captcha_id"`
	CaptchaAnswer string `json:"captcha_answer"`
}

// Register user baru
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	if len(c.Body()) == 0 {
		return utils.Error(c, fiber.StatusBadRequest, "Request body is required")
	}

	var body RegisterRequest
	if err := c.BodyParser(&body); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body: must be valid JSON")
	}

	// ===== Tambahkan verifikasi captcha di sini =====
	if body.CaptchaID == "" || body.CaptchaAnswer == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Captcha is required")
	}

	if !store.Store.Verify(body.CaptchaID, body.CaptchaAnswer) {
		return utils.Error(c, fiber.StatusBadRequest, "Captcha salah atau kadaluarsa")
	}
	// =================================================

	if body.Email == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Email is required")
	}
	if body.Password == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Password is required")
	}

	var exists bool
	if err := h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", body.Email).Scan(&exists); err != nil {
		log.Printf("Register: failed to check email existence: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "Database error")
	}
	if exists {
		return utils.Error(c, fiber.StatusConflict, "Email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Register: failed to hash password: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to create user")
	}

	var id int
	err = h.DB.QueryRow(
		"INSERT INTO users (email, password, created_at) VALUES ($1, $2, $3) RETURNING id",
		body.Email, string(hashedPassword), time.Now(),
	).Scan(&id)
	if err != nil {
		log.Printf("Register: failed to insert user: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to create user")
	}

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

	// ===== Tambahkan verifikasi captcha di sini =====
	if body.CaptchaID == "" || body.CaptchaAnswer == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Captcha is required")
	}

	if !store.Store.Verify(body.CaptchaID, body.CaptchaAnswer) {
		return utils.Error(c, fiber.StatusBadRequest, "Captcha salah atau kadaluarsa")
	}
	// =================================================

	if body.Email == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Email is required")
	}
	if body.Password == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Password is required")
	}

	var id int
	var email string
	var hashedPassword string
	err := h.DB.QueryRow("SELECT id, email, password FROM users WHERE email=$1", body.Email).Scan(&id, &email, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.Error(c, fiber.StatusUnauthorized, "Invalid email or password")
		}
		log.Printf("Login: failed to query user: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "Database error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(body.Password)); err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "Invalid email or password")
	}

	accessToken, err := utils.CreateAccessToken(id)
	if err != nil {
		log.Printf("Login: failed to create access token: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to create access token")
	}

	refreshToken, err := utils.CreateRefreshToken(id)
	if err != nil {
		log.Printf("Login: failed to create refresh token: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to create refresh token")
	}

	// Set cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: false,
		Path:     "/",
		MaxAge:   3600, // 1 jam

		SameSite: "Lax", // dev mode
		Secure:   false, // Set true jika menggunakan HTTPS di production */
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: false,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 hari
		SameSite: "Lax",            // dev mode
		Secure:   false,            // Set true jika menggunakan HTTPS di production */
	})

	// Generate CSRF token setelah cookie auth
	csrfToken, err := middleware.GenerateCSRFToken()
	if err != nil {
		log.Printf("Login: failed to generate CSRF token: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to create CSRF token")
	}

	// Set CSRF token cookie (bisa dibaca frontend)
	c.Cookie(&fiber.Cookie{
		Name:     middleware.CSRFCookieName,
		Value:    csrfToken,
		Path:     "/",
		HTTPOnly: false, // frontend perlu baca
		SameSite: "Lax",
		Secure:   false, // true jika production
	})

	// Response success (tanpa token di body)
	return utils.SuccessMessage(c, "Login successful", map[string]interface{}{
		"user": map[string]interface{}{
			"id":    id,
			"email": email,
		},
	}, nil)
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return utils.Error(c, fiber.StatusUnauthorized, "Refresh token missing")
	}

	userID, err := utils.ValidateRefreshToken(refreshToken)
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

	/// Generate new CSRF token
	csrfToken, err := middleware.GenerateCSRFToken()
	if err != nil {
		log.Printf("RefreshToken: failed to generate CSRF token: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "Failed to create CSRF token")
	}

	// Update cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Path:     "/",
		MaxAge:   3600,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HTTPOnly: true,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
	})

	// Set CSRF token cookie (bisa dibaca frontend)
	c.Cookie(&fiber.Cookie{
		Name:     middleware.CSRFCookieName,
		Value:    csrfToken,
		Path:     "/",
		HTTPOnly: false, // frontend perlu baca
		SameSite: "Lax",
		Secure:   false, // true jika production
	})

	return utils.SuccessMessage(c, "Token refreshed successfully", nil, nil)
}

// Logout digunakan untuk endpoint logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Clear cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		HTTPOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})

	// Clear CSRF token
	c.Cookie(&fiber.Cookie{
		Name:     middleware.CSRFCookieName,
		Value:    "",
		Path:     "/",
		HTTPOnly: false, // karena sebelumnya bisa dibaca frontend
		MaxAge:   -1,
	})

	return utils.SuccessMessage(c, "Logout successful", nil, nil)
}
