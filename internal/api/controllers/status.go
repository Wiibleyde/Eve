package controllers

import (
	"time"

	"Eve/internal/api/models"

	"github.com/gofiber/fiber/v3"
)

type StatusController struct{}

func (StatusController) Get(c fiber.Ctx) error {
	return c.JSON(models.StatusResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	})
}
