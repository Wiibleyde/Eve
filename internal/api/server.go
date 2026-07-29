package api

import (
	"Eve/internal/api/controllers"
	"Eve/internal/api/middleware"
	"Eve/internal/database"
	"Eve/internal/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func Start(port string, db *database.Client) {
	DB = db

	app := fiber.New(fiber.Config{
		AppName:        "Eve API",
		ReadBufferSize: 16 * 1024,
	})

	app.Use(middleware.RequestLogger())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{fiber.MethodGet},
		AllowCredentials: false,
	}))

	status := controllers.StatusController{}
	loto := controllers.LotoController{}

	v1 := app.Group("/api/v1")
	v1.Get("/status", status.Get)
	v1.Get("/health", status.Health)
	v1.Get("/info", status.Info)
	v1.Get("/ping", status.Ping)
	v1.Get("/loto/stats", loto.GetStats)
	v1.Get("/loto/winners", loto.GetWinners)

	logger.Info("API listening", "port", port)
	if err := app.Listen(":"+port, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
		logger.Fatal("API server error", "error", err)
	}
}
