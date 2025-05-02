package main

import (
	"main/pkg/bot"
	"main/pkg/bot_utils"
	"main/pkg/config"
	"main/pkg/intelligence"
	"main/pkg/logger"
)

func main() {
	bot_utils.InitStartTime()
	config.InitConfig()
	logger.InitLogger()

	logger.InfoLogger.Println("Program starting...")

	intelligence.InitAi()

	bot.InitBot()
}
