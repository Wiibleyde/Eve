package crons

import (
	"main/pkg/bot_utils"
	"main/pkg/data"
	"main/pkg/logger"
	"main/prisma/db"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func checkBirthdays(bot *discordgo.Session) {
	logger.InfoLogger.Println("Checking birthdays...")
	client, ctx := data.GetDBClient()

	// Get current month and day
	currentTime := time.Now()
	currentMonth := currentTime.Month()
	currentDay := currentTime.Day()

	// Get all users from database
	allUsers, err := client.GlobalUserData.FindMany().Exec(ctx)
	if err != nil {
		logger.ErrorLogger.Println("Error fetching users:", err)
		return
	}

	// Filter users whose birthdays are today
	var todayBirthdays []db.GlobalUserDataModel
	for _, user := range allUsers {
		birthDate, ok := user.BirthDate()
		if !ok || birthDate.IsZero() {
			continue
		}

		birthMonth := birthDate.Month()
		birthDay := birthDate.Day()

		if birthMonth == currentMonth && birthDay == currentDay {
			todayBirthdays = append(todayBirthdays, user)
		}
	}

	botGuilds := bot.State.Guilds
	for _, guild := range botGuilds {
		guildID := guild.ID
		for _, birthday := range todayBirthdays {
			userID := birthday.UserID
			member, err := bot.GuildMember(guildID, userID)
			if err != nil {
				logger.ErrorLogger.Println("Error fetching member:", err)
				continue
			}
			if member == nil {
				continue
			}
			channel, err := bot.UserChannelCreate(userID)
			if err != nil {
				logger.ErrorLogger.Println("Error creating channel:", err)
				continue
			}
			birthDate, ok := birthday.BirthDate()
			if !ok {
				logger.ErrorLogger.Println("Error getting birth date for user:", userID)
				continue
			}
			nowAge := time.Now().Year() - birthDate.Year()
			embed := bot_utils.BasicEmbedBuilder(bot)
			embed.Description = "Souhaitez un joyeux anniversaire à <@" + member.User.ID + "> (" + strconv.Itoa(nowAge) + " ans) 🎉"
			bot.ChannelMessageSendEmbed(channel.ID, embed)
		}
	}
}
