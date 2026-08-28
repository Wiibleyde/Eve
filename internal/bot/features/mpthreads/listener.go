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

const maxConcurrentRelays = 8

var relaySlots = make(chan struct{}, maxConcurrentRelays)

func Attach(client *bot.Client) {
	if !Enabled() {
		logger.Warn("MP threads: feature disabled, DM bridge not attached")
		return
	}

	if err := initWithRetry(context.Background()); err != nil {
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
	cfg, ok := bridgeSettings()
	if !ok {
		return
	}

	msg := e.Message
	client := e.Client()

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

func relayableType(t discord.MessageType) bool {
	switch t {
	case discord.MessageTypeDefault, discord.MessageTypeReply:
		return true
	default:
		return false
	}
}

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
