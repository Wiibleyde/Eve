package api

import (
	"Eve/internal/api/controllers"
	"Eve/internal/api/middleware"
	"Eve/internal/database"
	"Eve/internal/logger"

	"github.com/gofiber/fiber/v3"
)

func Start(port string, db *database.Client) {
	DB = db

	app := fiber.New(fiber.Config{
		AppName:        "Eve API",
		ReadBufferSize: 16 * 1024,
	})

	app.Use(middleware.RequestLogger())

	status := controllers.StatusController{}

	v1 := app.Group("/api/v1")
	v1.Get("/status", status.Get)

	logger.Info("API listening", "port", port)
	if err := app.Listen(":"+port, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
		logger.Fatal("API server error", "error", err)
	}
}
