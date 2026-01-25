package api

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/handler"
	"github.com/qwerius/gonuxt/internal/middleware"
)

func RegisterRoutes(app *fiber.App, db *sql.DB) {
	// Handlers
	userHandler := handler.NewUserHandler(db)
	authHandler := handler.NewAuthHandler(db)
   

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

	// User routes (protected)
	api.Get("/users", middleware.AuthRequired, userHandler.GetAllUsers)
	// nanti bisa tambah: api.Get("/users/:id", middleware.AuthRequired, userHandler.GetUserByID)
	// api.Post("/users", middleware.AuthRequired, userHandler.CreateUser)
	// api.Put("/users/:id", middleware.AuthRequired, userHandler.UpdateUser)
	// api.Delete("/users/:id", middleware.AuthRequired, userHandler.DeleteUser)


}
