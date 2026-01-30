package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/config"
	"github.com/qwerius/gonuxt/internal/utils"
)

type OAuthHandler struct {
	DB *sql.DB
}

func NewOAuthHandler(db *sql.DB) *OAuthHandler {
	return &OAuthHandler{DB: db}
}

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
	googleUserURL  = "https://www.googleapis.com/oauth2/v2/userinfo"
)

type GoogleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type GoogleUserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// ---------------------------
// GET /oauth/google/login
// ---------------------------
func (h *OAuthHandler) GoogleLogin(c *fiber.Ctx) error {
	params := url.Values{}
	params.Add("client_id", config.Get("GOOGLE_CLIENT_ID"))
	params.Add("redirect_uri", config.Get("GOOGLE_REDIRECT_URI"))
	params.Add("response_type", "code")
	params.Add("scope", "openid email profile")
	params.Add("access_type", "offline")
	params.Add("prompt", "consent")

	redirectURL := fmt.Sprintf("%s?%s", googleAuthURL, params.Encode())
	return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
}

// ---------------------------
// GET /oauth/google/callback
// ---------------------------
func (h *OAuthHandler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return utils.Error(c, fiber.StatusBadRequest, "code is required")
	}

	// 1. Exchange code -> token
	tokenResp, err := exchangeCodeToToken(code)
	if err != nil {
		log.Printf("GoogleCallback token exchange: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to exchange token")
	}

	// 2. Get user info
	userInfo, err := fetchGoogleUser(tokenResp.AccessToken)
	if err != nil {
		log.Printf("GoogleCallback fetch user: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to fetch user info")
	}

	// 3. Register or login
	userID, err := h.createOrGetUser(userInfo)
	if err != nil {
		log.Printf("GoogleCallback createOrGetUser: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to register/login user")
	}

	// 4. Generate JWT for your app
	jwtToken, err := utils.CreateAccessToken(userID)
	if err != nil {
		log.Printf("GoogleCallback generate jwt: %v", err)
		return utils.Error(c, fiber.StatusInternalServerError, "failed to generate token")
	}

	// 5. Return token to frontend
	return utils.SuccessMessage(c, "Login successful", map[string]string{
		"access_token": jwtToken,
		"token_type":   "Bearer",
	}, nil, nil)
}

// Exchange code -> token
func exchangeCodeToToken(code string) (*GoogleTokenResponse, error) {
	values := url.Values{}
	values.Add("code", code)
	values.Add("client_id", config.Get("GOOGLE_CLIENT_ID"))
	values.Add("client_secret", config.Get("GOOGLE_CLIENT_SECRET"))
	values.Add("redirect_uri", config.Get("GOOGLE_REDIRECT_URI"))
	values.Add("grant_type", "authorization_code")

	resp, err := http.PostForm(googleTokenURL, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

// Fetch user info
func fetchGoogleUser(accessToken string) (*GoogleUserResponse, error) {
	req, _ := http.NewRequest("GET", googleUserURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userResp GoogleUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, err
	}
	return &userResp, nil
}

// Register or get existing user
func (h *OAuthHandler) createOrGetUser(user *GoogleUserResponse) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id int
	err := h.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", user.Email).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	// user already exists
	if id != 0 {
		return id, nil
	}

	// create new user
	err = h.DB.QueryRowContext(ctx, `
		INSERT INTO users (email, created_at, updated_at)
		VALUES ($1, NOW(), NOW())
		RETURNING id
	`, user.Email).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}
