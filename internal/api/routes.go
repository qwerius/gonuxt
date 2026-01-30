package api

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/handler"
	"github.com/qwerius/gonuxt/internal/middleware"
)

func RegisterRoutes(app *fiber.App, db *sql.DB) {

	app.Use(middleware.CORS())

	// Handlers
	userHandler := handler.NewUserHandler(db)
	authHandler := handler.NewAuthHandler(db)
	roleHandler := handler.NewRoleHandler(db)
	userRoleHandler := handler.NewUserRoleHandler(db)
	profileHandler := handler.NewProfileHandler(db)
	oauthHandler := handler.NewOAuthHandler(db)

	// Root info
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":      "ok",
			"api_name":    "MyProject API",
			"version":     "1.0.0",
			"endpoints":   []string{"/hello", "/apa", "/api/v1/users", "/api/v1/auth/login", "/api/v1/auth/register"},
			"dokumentasi": "https://blueink.my.id",
		})
	})

	// Simple non-API routes
	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.SendString("Hello Fiber")
	})
	app.Get("/apa", func(c *fiber.Ctx) error {
		return c.SendString("Apa tanya-tanya")
	})

	// API v1 group
	api := app.Group("/api/v1")

	// Auth routes (no middleware)
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/refresh", authHandler.RefreshToken)
	api.Post("/auth/forgot-password", handler.ForgotPassword(db))
	api.Post("/auth/reset-password", handler.ResetPassword(db))

	app.Get("/oauth/google/login", oauthHandler.GoogleLogin)
	app.Get("/oauth/google/callback", oauthHandler.GoogleCallback)

	// User routes (protected)
	api.Get("/users", middleware.AuthRequired, userHandler.GetAllUsers)
	api.Get("/users/:id", middleware.AuthRequired, userHandler.GetUserByID)
	api.Post("/users", middleware.AuthRequired, userHandler.CreateUser)
	api.Put("/users/:id", middleware.AuthRequired, userHandler.UpdateUser)
	api.Delete("/users/:id", middleware.AuthRequired, userHandler.DeleteUser)

	api.Get("/roles", roleHandler.GetAllRoles)
	api.Get("/roles/:id", roleHandler.GetRoleByID)
	api.Delete("/roles/:id", roleHandler.DeleteRole)
	api.Post("/roles", roleHandler.CreateRole)

	api.Get("/users/:id/roles", userRoleHandler.GetUserRoles)
	api.Put("/users/:id/role", userRoleHandler.UpdateUserRole)
	api.Post("/users/:id/roles", userRoleHandler.AssignRole)
	api.Delete("/users/:id/roles/:roleId", userRoleHandler.RemoveRole)

	api.Get("/profiles/:id", middleware.AuthRequired, profileHandler.GetProfileByID)
	api.Get("/profiles", middleware.AuthRequired, profileHandler.GetAllProfiles)
	api.Get("/profile", middleware.AuthRequired, profileHandler.GetMyProfile)
	// ----------------------------
	// Profile by user_id
	// ----------------------------
	api.Get("/users/:id/profile", middleware.AuthRequired, profileHandler.GetProfileByUserID)
	api.Post("/users/:id/profile", middleware.AuthRequired, profileHandler.CreateProfileByUserID)
	api.Put("/users/:id/profile", middleware.AuthRequired, profileHandler.UpdateProfileByUserID)
	api.Delete("/users/:id/profile", middleware.AuthRequired, profileHandler.DeleteProfileByUserID)

}
