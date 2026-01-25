package api

import (
    "database/sql"

    "github.com/gofiber/fiber/v2"
    "github.com/qwerius/gonuxt/internal/handler"

)

func RegisterRoutes(app *fiber.App, db *sql.DB) {
    userHandler := handler.NewUserHandler(db)
     app.Get("/", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{    
            "status": "ok",       
            "api_name":  "MyProject API",
            "version":   "1.0.0",
            "endpoints": []string{"/hello", "/apa"},
            "dokumentasi": "https://blueink.my.id",
        })
    }) 
   /*  app.Get("/", func(c *fiber.Ctx) error {
        return c.Redirect("https://blueink.my.id")
        })
   */
    app.Get("/hello", func(c *fiber.Ctx) error {
        return c.SendString("Hello Fiber")
    })
	app.Get("/apa", func(c *fiber.Ctx) error {
		return c.SendString("Apa tanya-tanya")
	})

    app.Get("/users", userHandler.GetAllUsers)
}
