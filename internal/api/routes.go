// Package api berisi semua route dan pendaftaran middleware untuk API MyProject.
package api

import (
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/handler"
	"github.com/qwerius/gonuxt/internal/middleware"
)

func RegisterRoutes(app *fiber.App, db *sql.DB) {

	app.Use(middleware.CORS())
	app.Use(middleware.CSRF())

	// Handlers
	userHandler := handler.NewUserHandler(db)
	authHandler := handler.NewAuthHandler(db)
	roleHandler := handler.NewRoleHandler(db)
	userRoleHandler := handler.NewUserRoleHandler(db)
	profileHandler := handler.NewProfileHandler(db)
	oauthHandler := handler.NewOAuthHandler(db)
	auditHandler := handler.NewAuditHandler(db)
	captchaHandler := handler.NewCaptchaHandler()

	ipCfg := &middleware.IPFilterConfig{
		Whitelist: []string{"127.0.0.1", "192.168.1.0/24"},
		Blacklist: []string{"1.2.3.4"},
	}
	app.Use(middleware.IPFilterMiddleware(ipCfg))

	// Root info
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":      "ok",
			"api_name":    "MyProject API",
			"version":     "1.0.0",
			"dokumentasi": "https://blueink.my.id",
		})
	})

	api := app.Group("/api/v1")
	users := api.Group("/users", middleware.AuthRequired,
		middleware.RateLimit(middleware.RateLimitConfig{
			Max:        30,
			Expiration: time.Minute,
		}))

	authLimit := middleware.RateLimit(middleware.RateLimitConfig{
		Max:        5,
		Expiration: time.Minute,
	})

	api.Get("/captcha", captchaHandler.GenerateCaptcha)

	api.Post("/auth/login", authLimit, authHandler.Login)
	api.Post("/auth/register", authLimit, authHandler.Register)
	api.Post("/auth/refresh", authLimit, authHandler.RefreshToken)
	api.Post("/auth/forgot-password", authLimit, handler.ForgotPassword(db))
	api.Post("/auth/reset-password", authLimit, handler.ResetPassword(db))

	api.Get("/oauth/google/login", oauthHandler.GoogleLogin)
	api.Get("/oauth/google/callback", oauthHandler.GoogleCallback)

	users.Get("/", userHandler.GetAllUsers)
	users.Get("/:id", userHandler.GetUserByID)
	users.Post("/", middleware.AuthRequired, userHandler.CreateUser)
	users.Put("/:id", middleware.AuthRequired, userHandler.UpdateUser)
	users.Delete("/:id", middleware.AuthRequired, userHandler.DeleteUser)

	api.Get("/roles", middleware.AuthRequired, middleware.AdminOnly(db), roleHandler.GetAllRoles)
	api.Get("/roles/:id", middleware.AuthRequired, middleware.AdminOnly(db), roleHandler.GetRoleByID)
	api.Delete("/roles/:id", middleware.AuthRequired, middleware.AdminOnly(db), roleHandler.DeleteRole)
	api.Post("/roles", middleware.AuthRequired, middleware.AdminOnly(db), roleHandler.CreateRole)

	api.Get("/users/:id/roles", middleware.AuthRequired, middleware.AdminOnly(db), userRoleHandler.GetUserRoles)
	api.Put("/users/:id/role", middleware.AuthRequired, middleware.AdminOnly(db), userRoleHandler.UpdateUserRole)
	api.Post("/users/:id/roles", middleware.AuthRequired, middleware.AdminOnly(db), userRoleHandler.AssignRole)
	api.Delete("/users/:id/roles/:roleId", middleware.AuthRequired, middleware.AdminOnly(db), userRoleHandler.RemoveRole)
	api.Get("/role", middleware.AuthRequired, middleware.AuthRequired, roleHandler.GetMyRole)

	api.Get("/profiles/:id", middleware.AuthRequired, middleware.AdminOnly(db), profileHandler.GetProfileByID)
	api.Get("/profiles", middleware.AuthRequired, profileHandler.GetAllProfiles)
	api.Get("/profile", middleware.AuthRequired, profileHandler.GetMyProfile)

	api.Get("/users/:id/profile", middleware.AuthRequired, profileHandler.GetProfileByUserID)
	api.Post("/users/:id/profile", middleware.AuthRequired, middleware.OwnerOrAdminMiddleware(), profileHandler.CreateProfileByUserID)
	api.Put("/users/:id/profile", middleware.AuthRequired, middleware.OwnerOrAdminMiddleware(), profileHandler.UpdateProfileByUserID)
	api.Delete("/users/:id/profile", middleware.AuthRequired, middleware.OwnerOrAdminMiddleware(), profileHandler.DeleteProfileByUserID)

	api.Get("/admin/profile/:id", middleware.AuthRequired, profileHandler.GetProfileByAdmin)

	auditCfg := &middleware.AuditConfig{DB: db}

	api.Get("/audit-logs",
		middleware.AuthRequired,                     // pastikan user ada di context
		middleware.AdminOnly(db),                    // harus admin
		middleware.AuditLoggingMiddleware(auditCfg), // catat audit log
		auditHandler.GetAuditLogs,                   // handler untuk menampilkan log
	)

}
