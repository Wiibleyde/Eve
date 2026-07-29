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

// processStart is initialised when this package is loaded, i.e. during program
// start-up before main() runs. That is close enough to the real process start
// for an uptime counter and saves wiring a timestamp through from main.
var processStart = time.Now()

// StatusController serves the liveness / metadata endpoints: /status, /health,
// /info and /ping.
//
// The three liveness routes are near duplicates kept for compatibility with the
// previous TS API; serving them from a single controller keeps them from
// drifting apart.
type StatusController struct{}

// Get handles GET /api/v1/status.
func (StatusController) Get(c fiber.Ctx) error {
	return c.JSON(models.StatusResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	})
}

// Health handles GET /api/v1/health.
func (StatusController) Health(c fiber.Ctx) error {
	return c.JSON(models.HealthResponse{
		Health: "good",
		Uptime: time.Since(processStart).Seconds(),
	})
}

// Info handles GET /api/v1/info. The version is injected at build time through
// -ldflags, see package Eve/internal/version.
func (StatusController) Info(c fiber.Ctx) error {
	return c.JSON(models.InfoResponse{
		App:     appName,
		Version: version.Version,
		Author:  appAuthor,
	})
}

// Ping handles GET /api/v1/ping.
func (StatusController) Ping(c fiber.Ctx) error {
	return c.JSON(models.PingResponse{
		Message:   "pong",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
