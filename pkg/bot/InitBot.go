package bot

import (
	"main/pkg/config"
	"main/pkg/crons"
	"main/pkg/logger"
	"os"
	"os/signal"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
)

func InitBot() {
	logger.InfoLogger.Println("Démarrage du bot...")
	client, err := disgo.New(config.GetConfig().DiscordToken, bot.WithGatewayConfigOpts(
		gateway.WithIntents(
			gateway.IntentGuilds,
			gateway.IntentGuildMessages,
			gateway.IntentGuildMessageReactions,
			gateway.IntentGuildMessageTyping,
			gateway.IntentGuildVoiceStates,
			gateway.IntentGuildMembers,
			gateway.IntentDirectMessages,
			gateway.IntentDirectMessageReactions,
			gateway.IntentDirectMessageTyping,
			gateway.IntentMessageContent,
		),
		
	))
	if err != nil {
		logger.ErrorLogger.Panicln("[PANIC] Le bot n'a pas pu démarrer,", err)
		return
	}

	// dg.AddHandler(messageCreate)
	// dg.AddHandler(onReady)
	// dg.AddHandler(interactionCreate)

	err = dg.Open()
	if err != nil {
		logger.ErrorLogger.Panicln("Impossible d'ouvrir la connexion", err.Error())
		return
	}

	crons.StartCron(dg)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	logger.InfoLogger.Println("Réception du signal d'arrêt, fermeture du bot...")

	err = dg.Close()
	if err != nil {
		logger.InfoLogger.Printf("Impossible de fermer le bot : %v", err)
	}

	logger.InfoLogger.Println("Bot arrêté avec succès")
}
