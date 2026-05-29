package main

import (
	"Eve/internal/bot"
	"Eve/internal/logger"

	"github.com/joho/godotenv"
)

func main() {
	logger.Info("Bot starting...")

	godotenv.Load()

	bot.Run()
}
