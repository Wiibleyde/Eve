package controllers

import (
	"time"

	"Eve/internal/api/models"
	"Eve/internal/version"

	"github.com/gofiber/fiber/v3"
)

const (
	appName   = "Eve - API"
	appAuthor = "Wiibleyde"
)

var processStart = time.Now()

type StatusController struct{}

func (StatusController) Get(c fiber.Ctx) error {
	return c.JSON(models.StatusResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	})
}

func (StatusController) Health(c fiber.Ctx) error {
	return c.JSON(models.HealthResponse{
		Health: "good",
		Uptime: time.Since(processStart).Seconds(),
	})
}

func (StatusController) Info(c fiber.Ctx) error {
	return c.JSON(models.InfoResponse{
		App:     appName,
		Version: version.Version,
		Author:  appAuthor,
	})
}

func (StatusController) Ping(c fiber.Ctx) error {
	return c.JSON(models.PingResponse{
		Message:   "pong",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
