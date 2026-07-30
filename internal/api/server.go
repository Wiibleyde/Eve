package api

import (
	"net/http"

	"Eve/internal/api/controllers"
	"Eve/internal/api/middleware"
	"Eve/internal/api/models"
	"Eve/internal/database"
	"Eve/internal/logger"
	"Eve/internal/version"

	"github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

type apiResponse struct {
	status      int
	body        any
	description string
}

type apiRoute struct {
	path      string
	handler   fiber.Handler
	tag       string
	id        string
	summary   string
	query     any
	responses []apiResponse
}

func routes(status controllers.StatusController, loto controllers.LotoController) []apiRoute {
	dbErrors := []apiResponse{
		{http.StatusInternalServerError, controllers.APIError{}, "A database query failed"},
		{http.StatusServiceUnavailable, controllers.APIError{}, "The database client is not initialised"},
	}

	return []apiRoute{
		{
			path: "/api/v1/status", handler: status.Get, tag: "status", id: "getStatus",
			summary: "Static ok marker plus the current server time",
			responses: []apiResponse{
				{http.StatusOK, models.StatusResponse{}, "Service is up"},
			},
		},
		{
			path: "/api/v1/health", handler: status.Health, tag: "status", id: "getHealth",
			summary: "Health marker plus the process uptime in seconds",
			responses: []apiResponse{
				{http.StatusOK, models.HealthResponse{}, "Health report"},
			},
		},
		{
			path: "/api/v1/info", handler: status.Info, tag: "status", id: "getInfo",
			summary: "Application name, build version and author",
			responses: []apiResponse{
				{http.StatusOK, models.InfoResponse{}, "Application information"},
			},
		},
		{
			path: "/api/v1/ping", handler: status.Ping, tag: "status", id: "getPing",
			summary: "Pong plus the current server time as RFC 3339",
			responses: []apiResponse{
				{http.StatusOK, models.PingResponse{}, "Pong"},
			},
		},
		{
			path: "/api/v1/loto/stats", handler: loto.GetStats, tag: "loto", id: "getLotoStats",
			summary: "Per-game loto statistics plus aggregated totals, newest first",
			query:   statsQuery{},
			responses: append([]apiResponse{
				{http.StatusOK, controllers.LotoStatsResponse{}, "Statistics for the matching games, empty when nothing matches"},
			}, dbErrors...),
		},
		{
			path: "/api/v1/loto/winners", handler: loto.GetWinners, tag: "loto", id: "getLotoWinners",
			summary: "Drawn loto prizes that have a winner, newest first",
			query:   winnersQuery{},
			responses: append([]apiResponse{
				{http.StatusOK, controllers.LotoWinnersResponse{}, "Matching winners, empty when nothing matches"},
				{http.StatusBadRequest, controllers.APIError{}, "The limit parameter is not a positive integer"},
			}, dbErrors...),
		},
	}
}

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

	endpoints := routes(controllers.StatusController{}, controllers.LotoController{})

	if spec, err := buildSpec(version.Version, endpoints); err != nil {
		logger.Error("API docs disabled", "error", err)
	} else {
		app.Use(swaggerui.New(swaggerui.Config{
			BasePath:    "/",
			FilePath:    specPath,
			FileContent: spec,
			Path:        uiPath,
			Title:       specTitle,
		}))
	}

	for _, route := range endpoints {
		app.Get(route.path, route.handler)
	}

	logger.Info("API listening", "port", port, "docs", "/"+uiPath)
	if err := app.Listen(":"+port, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
		logger.Fatal("API server error", "error", err)
	}
}
