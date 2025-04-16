package bot

import (
	"main/pkg/config"
	"main/pkg/data"
	"main/pkg/logger"
	"main/prisma/db"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
)

func InitBot() {
	logger.InfoLogger.Println("Starting bot...", config.GetConfig().DiscordToken)
	dg, err := discordgo.New("Bot " + config.GetConfig().DiscordToken)
	if err != nil {
		logger.ErrorLogger.Panicln("[PANIC] Error creating Discord session,", err)
		return
	}

	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions | discordgo.IntentsGuildVoiceStates

	dg.AddHandler(MessageCreate)
	dg.AddHandler(OnReady)
	dg.AddHandler(InteractionCreate)

	err = dg.Open()
	if err != nil {
		logger.ErrorLogger.Panicln("Error opening connection,", err.Error())
		return
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	err = dg.Close()
	if err != nil {
		logger.InfoLogger.Printf("could not close session gracefully: %s", err)
	}
}

func CheckBirthdays(bot *discordgo.Session) {
	logger.InfoLogger.Println("Checking birthdays...")
	client, ctx := data.GetDBClient()
	today := time.Now().Format("2006-01-02")
	todayDate, _ := time.Parse("2006-01-02", today)
	todayBirthdays, err := client.GlobalUserData.FindMany(
		db.GlobalUserData.BirthDate.Equals(db.DateTime(todayDate)),
	).Exec(ctx)
	if err != nil {
		logger.ErrorLogger.Println("Error fetching birthdays:", err)
		return
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
			embed := &discordgo.MessageEmbed{
				Title:       "Joyeux Anniversaire !",
				Description: "Aujourd'hui c'est ton anniversaire ! 🎉",
				Color:       0x00FF00,
			}
			bot.ChannelMessageSendEmbed(channel.ID, embed)
		}
	}

}
