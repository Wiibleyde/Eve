package bot

import (
	"main/pkg/config"
	"main/pkg/logger"
	"os"
	"os/signal"

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
