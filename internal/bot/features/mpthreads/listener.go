package mpthreads

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

const (
	initAttempts = 3
	initBackoff  = time.Second
)

// maxConcurrentRelays bounds the goroutines downloading attachments and calling
// the REST API at once; each of them buffers up to maxUploadBytes in memory.
const maxConcurrentRelays = 8

var relaySlots = make(chan struct{}, maxConcurrentRelays)

// Attach loads the persisted mappings and registers the DM ↔ thread bridge on
// the client. It must be called before OpenGateway and is a no-op when the
// feature is not configured.
//
// Required intents: gateway.IntentDirectMessages, gateway.IntentGuildMessages
// and gateway.IntentMessageContent.
func Attach(client *bot.Client) {
	if !Enabled() {
		logger.Warn("MP threads: feature disabled, DM bridge not attached")
		return
	}

	if err := initWithRetry(context.Background()); err != nil {
		// The DB is the source of truth. Serving DMs from a knowingly empty cache
		// would open a second thread for every user who already has one, and each
		// of those mappings would then fight the user_id unique index.
		logger.Error("MP threads: loading thread mappings failed, DM bridge not attached", "error", err)
		return
	}

	client.AddEventListeners(bot.NewListenerFunc(onMessageCreate))
	logger.Info("MP threads: DM bridge attached", "mappings", MappingCount())
}

func initWithRetry(ctx context.Context) error {
	var err error
	for attempt := range initAttempts {
		if err = Init(ctx); err == nil {
			return nil
		}
		if attempt < initAttempts-1 {
			logger.Warn("MP threads: loading thread mappings failed, retrying",
				"attempt", attempt+1, "error", err)
			time.Sleep(initBackoff * time.Duration(attempt+1))
		}
	}
	return err
}

func onMessageCreate(e *events.MessageCreate) {
	cfg, ok := config()
	if !ok {
		return
	}

	msg := e.Message
	client := e.Client()

	// Loop safety: never relay anything the bot itself (or a webhook) wrote.
	if msg.Author.Bot || msg.Author.System || msg.WebhookID != nil || msg.Author.ID == client.ID() {
		return
	}
	if !relayableType(msg.Type) {
		return
	}

	if e.GuildID == nil {
		dispatch("inbound", msg.Author.ID, func() { relayInbound(client, msg) })
		return
	}
	if *e.GuildID != cfg.guildID {
		return
	}
	userID, ok := userForThread(msg.ChannelID)
	if !ok {
		return
	}
	dispatch("outbound", userID, func() { relayOutbound(client, msg, userID) })
}

// relayableType keeps system messages (pins, joins, thread starters…) out of
// the bridge.
func relayableType(t discord.MessageType) bool {
	switch t {
	case discord.MessageTypeDefault, discord.MessageTypeReply:
		return true
	default:
		return false
	}
}

// dispatch runs a relay off the gateway loop — it performs HTTP downloads and
// several REST calls — with panic isolation so a malformed message can never
// take the bot down.
//
// Everything about one conversation runs under that conversation's lock, in
// both directions, so a burst of DMs or of staff replies reaches the other side
// in the order it was written. The slot is taken after the lock and never
// before: a goroutine holding a slot already owns its lock and can therefore
// never wait on one, which is what keeps the two from deadlocking. Nothing is
// dropped on the way — a lost support message would be worse than a queued one.
func dispatch(direction string, userID snowflake.ID, fn func()) {
	go func() {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			logger.Error("MP threads: panic while relaying",
				"direction", direction,
				"panic", fmt.Sprint(rec),
				"stack", strings.ReplaceAll(string(debug.Stack()), "\n", " | "),
			)
		}()

		unlock := lockUser(userID)
		defer unlock()

		relaySlots <- struct{}{}
		defer func() { <-relaySlots }()

		fn()
	}()
}
