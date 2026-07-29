package reactions

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"Eve/internal/bot/maintenance"
	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/guildconfig"
	"Eve/internal/database/tables"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

const lookupTimeout = 3 * time.Second

func Attach(client *bot.Client) {
	client.AddEventListeners(bot.NewListenerFunc(onMessageCreate))
}

func onMessageCreate(e *events.MessageCreate) {
	if e.GuildID == nil {
		return
	}
	msg := e.Message
	if msg.Author.Bot || msg.Author.System || msg.WebhookID != nil {
		return
	}
	if maintenance.Enabled() {
		return
	}
	if mentionsSelf(e) {
		return
	}

	reply, ok := Detect(msg.Content)
	if !ok {
		return
	}
	if !cooldowns.allow(e.ChannelID, time.Now()) {
		logger.Debug("Reaction reply suppressed by cooldown", "channel", e.ChannelID.String())
		return
	}

	guildID := *e.GuildID
	client := e.Client()
	channelID := e.ChannelID
	authorID := msg.Author.ID

	go func() {
		defer recoverPanic("reaction reply")

		ctx, cancel := context.WithTimeout(context.Background(), lookupTimeout)
		defer cancel()

		if jokesDisabled(ctx, guildID) {
			logger.Debug("Reaction reply skipped, jokes disabled", "guild", guildID.String())
			return
		}
		send(client, channelID, authorID, reply)
	}()
}

func mentionsSelf(e *events.MessageCreate) bool {
	selfID := e.Client().ID()
	for _, user := range e.Message.Mentions {
		if user.ID == selfID {
			return true
		}
	}
	return false
}

func jokesDisabled(ctx context.Context, guildID snowflake.ID) bool {
	if database.Default == nil {
		logger.Warn("Reactions: database unavailable, staying silent")
		return true
	}

	cfg, err := database.Default.Ent().GuildConfig.Query().
		Where(
			guildconfig.GuildID(guildID.String()),
			guildconfig.Key(tables.JokesDisabled.String()),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false
		}
		logger.Error("Error reading jokes.disabled", "guild", guildID.String(), "error", err)
		return true
	}
	return parseBool(cfg.Value)
}

func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "oui", "yes", "on":
		return true
	default:
		return false
	}
}

func send(client *bot.Client, channelID snowflake.ID, authorID snowflake.ID, reply string) {
	if _, err := client.Rest.CreateMessage(channelID, discord.MessageCreate{
		Content:         reply,
		AllowedMentions: &discord.AllowedMentions{Parse: []discord.AllowedMentionType{}},
	}); err != nil {
		logger.Error("Error sending reaction reply", "channel", channelID.String(), "error", err)
		return
	}
	logger.Debug("Reaction reply sent",
		"channel", channelID.String(),
		"user", authorID.String(),
		"reply", reply,
	)
}

func recoverPanic(what string) {
	rec := recover()
	if rec == nil {
		return
	}
	logger.Error("Panic in reactions listener",
		"what", what,
		"panic", fmt.Sprint(rec),
		"stack", strings.ReplaceAll(string(debug.Stack()), "\n", " | "),
	)
}
