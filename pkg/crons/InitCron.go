package crons

import (
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
)

func StartCron(bot *discordgo.Session) {
	logger.InfoLogger.Println("Starting cron jobs...")

	// Create a new cron scheduler
	c := cron.New(cron.WithSeconds())
	// Schedule the birthday check every day at midnight
	_, err := c.AddFunc("0 0 0 * * *", func() {
		checkBirthdays(bot)
	})
	if err != nil {
		logger.ErrorLogger.Println("Error adding cron job:", err)
		return
	}
	c.Start()
}
