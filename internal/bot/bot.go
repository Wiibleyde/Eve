package bot

import (
	"Eve/internal/bot/embeds"
	"Eve/internal/bot/features/birthday"
	"Eve/internal/bot/features/blague"
	"Eve/internal/bot/features/calendar"
	"Eve/internal/bot/features/coinflip"
	configfeature "Eve/internal/bot/features/config"
	debugfeature "Eve/internal/bot/features/debug"
	"Eve/internal/bot/features/loto"
	maintenancefeature "Eve/internal/bot/features/maintenance"
	"Eve/internal/bot/features/motus"
	"Eve/internal/bot/features/mpthreads"
	"Eve/internal/bot/features/ping"
	"Eve/internal/bot/features/presence"
	"Eve/internal/bot/features/quiz"
	"Eve/internal/bot/features/quote"
	"Eve/internal/bot/features/reactions"
	"Eve/internal/bot/features/streamer"
	"Eve/internal/bot/features/talk"
	"Eve/internal/bot/features/userinfo"
	"Eve/internal/bot/helpers"
	"Eve/internal/bot/router"
	"Eve/internal/config"
	"Eve/internal/database"
	"Eve/internal/logger"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
)

var DB *database.Client

func Run(cfg *config.Config, db *database.Client) {
	DB = db
	database.Default = db
	client, err := disgo.New(cfg.DiscordToken,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMembers,
				gateway.IntentGuildMessages,
				gateway.IntentDirectMessages,
				gateway.IntentMessageContent,
			),
		),
	)
	if err != nil {
		logger.Fatal("Error creating Discord client", "error", err)
	}

	helpers.SetOwner(cfg.BotOwnerID)

	r := router.New()
	birthday.Register(r)
	blague.Register(r)
	calendar.Register(r)
	coinflip.Register(r)
	configfeature.Register(r)
	debugfeature.Register(r)
	loto.Register(r)
	maintenancefeature.Register(r)
	motus.Register(r)
	ping.Register(r)
	quiz.Register(r)
	quote.Register(r)
	streamer.Register(r)
	talk.Register(r)
	userinfo.Register(r)
	r.Attach(client)

	reactions.Attach(client)
	mpthreads.Attach(client)

	allCommands := []discord.ApplicationCommandCreate{}
	allCommands = append(allCommands, birthday.Commands...)
	allCommands = append(allCommands, blague.Commands()...)
	allCommands = append(allCommands, calendar.Commands...)
	allCommands = append(allCommands, coinflip.Commands...)
	allCommands = append(allCommands, configfeature.Commands...)
	allCommands = append(allCommands, debugfeature.Commands...)
	allCommands = append(allCommands, loto.Commands...)
	allCommands = append(allCommands, maintenancefeature.Commands()...)
	allCommands = append(allCommands, motus.Commands...)
	allCommands = append(allCommands, ping.Commands...)
	allCommands = append(allCommands, quiz.Commands...)
	allCommands = append(allCommands, quote.Commands...)
	allCommands = append(allCommands, streamer.Commands()...)
	allCommands = append(allCommands, talk.Commands()...)
	allCommands = append(allCommands, userinfo.Commands...)

	var schedulersOnce sync.Once

	client.AddEventListeners(bot.NewListenerFunc(func(e *events.Ready) {
		logger.Info("Logged in", "user", e.User.Username)
		embeds.Init(e.User.EffectiveAvatarURL())
		schedulersOnce.Do(func() {
			birthday.StartScheduler(client)
			calendar.StartScheduler(client)
			presence.StartScheduler(client)
			quiz.StartScheduler(client)
			streamer.StartPoller(client)
		})
		if _, err := client.Rest.SetGlobalCommands(e.Application.ID, allCommands); err != nil {
			logger.Error("Error registering commands", "error", err)
		} else {
			logger.Info("Commands registered", "count", len(allCommands))
		}
		go clearGuildCommands(client, e.Application.ID, e.Guilds)
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

func clearGuildCommands(client *bot.Client, appID snowflake.ID, guilds []discord.UnavailableGuild) {
	for _, g := range guilds {
		if _, err := client.Rest.SetGuildCommands(appID, g.ID, []discord.ApplicationCommandCreate{}); err != nil {
			logger.Warn("Error clearing guild commands", "guild", g.ID, "error", err)
		}
	}
	logger.Info("Stale guild commands cleared", "guilds", len(guilds))
}
