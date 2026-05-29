package bot

import (
	"Eve/internal/bot/embeds"
	"Eve/internal/bot/features/birthday"
	configfeature "Eve/internal/bot/features/config"
	"Eve/internal/bot/features/ping"
	"Eve/internal/bot/features/quote"
	"Eve/internal/bot/router"
	"Eve/internal/config"
	"Eve/internal/database"
	"Eve/internal/logger"
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
)

var DB *database.Client

func Run(cfg *config.Config, db *database.Client) {
	DB = db
	database.Default = db
	client, err := disgo.New(cfg.DiscordToken,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildMembers),
		),
	)
	if err != nil {
		logger.Fatal("Error creating Discord client", "error", err)
	}

	r := router.New()
	birthday.Register(r)
	configfeature.Register(r)
	ping.Register(r)
	quote.Register(r)
	r.Attach(client)

	allCommands := []discord.ApplicationCommandCreate{}
	allCommands = append(allCommands, birthday.Commands...)
	allCommands = append(allCommands, configfeature.Commands...)
	allCommands = append(allCommands, ping.Commands...)
	allCommands = append(allCommands, quote.Commands...)

	client.AddEventListeners(bot.NewListenerFunc(func(e *events.Ready) {
		logger.Info("Logged in", "user", e.User.Username)
		embeds.Init(e.User.EffectiveAvatarURL())
		birthday.StartScheduler(client)
		if _, err := client.Rest.SetGlobalCommands(e.Application.ID, allCommands); err != nil {
			logger.Error("Error registering commands", "error", err)
		} else {
			logger.Info("Commands registered", "count", len(allCommands))
		}
	}))

	ctx := context.Background()
	if err := client.OpenGateway(ctx); err != nil {
		logger.Fatal("Error opening connection", "error", err)
	}
	defer client.Close(ctx)

	logger.Info("Bot is running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Info("Shutting down...")
}
