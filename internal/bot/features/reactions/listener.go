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

// lookupTimeout caps the opt-out lookup: a joke is never worth holding a
// database connection for.
const lookupTimeout = 3 * time.Second

// Attach registers the MessageCreate listener. It must be called before
// OpenGateway, and requires the IntentGuildMessages and IntentMessageContent
// gateway intents — without the (privileged) message content intent, message
// contents arrive empty and nothing ever matches.
func Attach(client *bot.Client) {
	client.AddEventListeners(bot.NewListenerFunc(onMessageCreate))
}

// onMessageCreate runs the cheap, synchronous guards (the gateway dispatches
// listeners inline, so nothing slow belongs here) and hands the database
// lookup plus the REST call to a goroutine.
func onMessageCreate(e *events.MessageCreate) {
	if e.GuildID == nil {
		return // guild messages only, DMs belong to the MP bridge
	}
	msg := e.Message
	if msg.Author.Bot || msg.Author.System || msg.WebhookID != nil {
		return // never answer bots, and never answer ourselves
	}
	if maintenance.Enabled() {
		return // silent skip, a joke is not worth breaking a maintenance window
	}
	if mentionsSelf(e) {
		return // mentions belong to the AI/talk path
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

// mentionsSelf reports whether the bot itself is mentioned. Replies to a bot
// message count as a mention, which conveniently also breaks joke-on-joke
// loops.
func mentionsSelf(e *events.MessageCreate) bool {
	selfID := e.Client().ID()
	for _, user := range e.Message.Mentions {
		if user.ID == selfID {
			return true
		}
	}
	return false
}

// jokesDisabled reads the per-guild opt-out from guild_configs
// ("jokes.disabled"). It fails closed: when the database is unreachable the
// guild stays silent rather than risking replies in a guild that opted out.
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
			return false // unset means enabled
		}
		logger.Error("Error reading jokes.disabled", "guild", guildID.String(), "error", err)
		return true
	}
	return parseBool(cfg.Value)
}

// parseBool is tolerant on purpose: the value is written by /config, but a
// human editing the table by hand should not silently break the opt-out.
func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "oui", "yes", "on":
		return true
	default:
		return false
	}
}

// send posts the joke as a plain message in the channel. AllowedMentions is
// explicitly empty so a reply can never ping anyone, whatever a future joke
// contains.
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

// recoverPanic keeps a broken joke from taking the whole process down: a panic
// in a bare goroutine is fatal, unlike one in an interaction handler (the
// router already recovers those).
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
