package bot

import (
	"main/pkg/config"
	"main/pkg/crons"
	"main/pkg/logger"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var (
	session *discordgo.Session
)

func InitBot() {
	logger.InfoLogger.Println("Starting bot...")
	dg, err := discordgo.New("Bot " + config.GetConfig().DiscordToken)
	if err != nil {
		logger.ErrorLogger.Panicln("[PANIC] Error creating Discord session,", err)
		return
	}

	session = dg

	dg.StateEnabled = true

	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildMessageTyping |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsDirectMessageReactions |
		discordgo.IntentsDirectMessageTyping |
		discordgo.IntentsMessageContent

	dg.AddHandler(messageCreate)
	dg.AddHandler(onReady)
	dg.AddHandler(interactionCreate)

	err = dg.Open()
	if err != nil {
		logger.ErrorLogger.Panicln("Error opening connection,", err.Error())
		return
	}

	crons.StartCron(dg)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	logger.InfoLogger.Println("Received interrupt signal, shutting down gracefully...")

	err = dg.Close()
	if err != nil {
		logger.InfoLogger.Printf("could not close session gracefully: %s", err)
	}

	logger.InfoLogger.Println("Bot has been shut down.")
}
