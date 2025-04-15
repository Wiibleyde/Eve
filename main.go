package main

import (
	"main/pkg/bot"
	"main/pkg/config"
	"main/pkg/logger"
)

func main() {
	config.InitConfig()
	logger.InitLogger()
	logger.InfoLogger.Println("Program starting...")

	bot.InitBot()
}
