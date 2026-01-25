package main

import (
    "fmt"
    "log"

    "github.com/qwerius/gonuxt/internal/db"
	"github.com/qwerius/gonuxt/internal/api"
	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/config"
)

func main() {
	config.Load()
    database, err := db.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer database.Close() 

    fmt.Println("Connected to PostgreSQL successfully!")

	app := fiber.New()
    api.RegisterRoutes(app, database)

	port := config.Get("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
    
}