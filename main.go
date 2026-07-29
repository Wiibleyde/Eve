package main

import (
	"Eve/internal/api"
	"Eve/internal/bot"
	"Eve/internal/config"
	"Eve/internal/database"
	"Eve/internal/logger"
	"context"

	"github.com/joho/godotenv"
)

func main() {
	logger.Info("Bot starting...")

	godotenv.Load()

	cfg := config.Load()

	botDB, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Error connecting to bot database", "error", err)
	}
	if err := botDB.Migrate(context.Background()); err != nil {
		logger.Fatal("Error running migrations", "error", err)
	}

	apiDB, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Error connecting to API database", "error", err)
	}

	go api.Start(cfg.APIPort, apiDB)

	bot.Run(cfg, botDB)
}
