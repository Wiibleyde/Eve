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

const gifUrl = "https://media0.giphy.com/media/v1.Y2lkPTc5MGI3NjExMHdoaTB6a3E4cmo3NGNmYjd3a3RlbjQ4a2pycXFpZGdxaTVkcTF0byZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/JksahSdnH6BX1WEtlc/giphy.gif"

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

		// Get birthday channel configuration
		birthdayChannelConfig, err := client.Config.FindMany(
			db.Config.And(
				db.Config.GuildID.Equals(guildID),
				db.Config.Key.Equals("birthdayChannel"),
			),
		).Exec(ctx)

		// Skip if no channel is configured or there's an error
		if err != nil || len(birthdayChannelConfig) == 0 {
			logger.InfoLogger.Printf("No birthday channel configured for guild %s, skipping", guildID)
			continue
		}

		birthdayChannelID := birthdayChannelConfig[0].Value

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

			birthDate, ok := birthday.BirthDate()
			if !ok {
				logger.ErrorLogger.Println("Error getting birth date for user:", userID)
				continue
			}

			nowAge := time.Now().Year() - birthDate.Year()

			embed := bot_utils.BasicEmbedBuilder(bot)
			embed.Title = "Joyeux Anniversaire ! 🎂"
			embed.Description = "Souhaitez un joyeux anniversaire à <@" + member.User.ID + "> (" + strconv.Itoa(nowAge) + " ans) 🎉"
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: gifUrl,
			}

			_, err = bot.ChannelMessageSendEmbed(birthdayChannelID, embed)
			if err != nil {
				logger.ErrorLogger.Printf("Failed to send birthday message in channel %s: %v", birthdayChannelID, err)
			}
		}
	}
}
