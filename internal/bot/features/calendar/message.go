package calendar

import (
	"context"
	"errors"
	"fmt"

	"Eve/internal/database/tables"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// errCalendarMessageGone is returned when the stored calendar message (or its
// channel) no longer exists. The stored refs have already been cleared when
// this error surfaces, so callers only have to decide whether to repost.
var errCalendarMessageGone = errors.New("calendar message no longer exists")

var errNoCalendarMessage = errors.New("no calendar message stored")

// updateCalendarMessage edits the stored calendar message with a fresh embed.
//
// When Discord answers "unknown message" or "unknown channel", the message was
// deleted or the channel is gone: the stored refs are cleared, a warning is
// logged and errCalendarMessageGone is returned.
func updateCalendarMessage(ctx context.Context, client *bot.Client, cfg guildCalendar, embed discord.Embed) error {
	if !cfg.hasMessage() {
		return errNoCalendarMessage
	}

	channelID, err := snowflake.Parse(cfg.ChannelID)
	if err != nil {
		return fmt.Errorf("parsing calendar channel id %q: %w", cfg.ChannelID, err)
	}
	messageID, err := snowflake.Parse(cfg.MessageID)
	if err != nil {
		return fmt.Errorf("parsing calendar message id %q: %w", cfg.MessageID, err)
	}

	embedList := []discord.Embed{embed}
	if _, err := client.Rest.UpdateMessage(channelID, messageID, discord.MessageUpdate{
		Embeds: &embedList,
	}); err != nil {
		if isGone(err) {
			logger.Warn("Calendar message no longer exists, clearing stored refs",
				"guild", cfg.GuildID, "channel", cfg.ChannelID, "message", cfg.MessageID)
			clearMessageRefs(ctx, cfg)
			return errCalendarMessageGone
		}
		return fmt.Errorf("editing calendar message: %w", err)
	}
	return nil
}

// postCalendarMessage publishes a new calendar message and stores its refs.
// A message whose refs could not be stored is deleted again: nothing would ever
// refresh or clean it up otherwise.
func postCalendarMessage(ctx context.Context, client *bot.Client, guildID string, channelID snowflake.ID, embed discord.Embed) error {
	message, err := client.Rest.CreateMessage(channelID, discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
	if err != nil {
		return fmt.Errorf("posting calendar message: %w", err)
	}

	if err := storeMessageRefs(ctx, guildID, channelID, message.ID); err != nil {
		if deleteErr := client.Rest.DeleteMessage(channelID, message.ID); deleteErr != nil && !isGone(deleteErr) {
			logger.Warn("Calendar: could not delete the orphaned calendar message",
				"guild", guildID, "channel", channelID.String(), "message", message.ID.String(), "error", deleteErr)
		}
		return err
	}
	return nil
}

func storeMessageRefs(ctx context.Context, guildID string, channelID snowflake.ID, messageID snowflake.ID) error {
	if err := saveConfig(ctx, guildID, tables.CalendarChannel, channelID.String()); err != nil {
		return err
	}
	return saveConfig(ctx, guildID, tables.CalendarMessage, messageID.String())
}

// deleteCalendarMessage removes the stored calendar message, best effort.
func deleteCalendarMessage(client *bot.Client, cfg guildCalendar) {
	if !cfg.hasMessage() {
		return
	}
	channelID, err := snowflake.Parse(cfg.ChannelID)
	if err != nil {
		return
	}
	messageID, err := snowflake.Parse(cfg.MessageID)
	if err != nil {
		return
	}
	if err := client.Rest.DeleteMessage(channelID, messageID); err != nil && !isGone(err) {
		logger.Warn("Calendar: could not delete calendar message",
			"guild", cfg.GuildID, "channel", cfg.ChannelID, "message", cfg.MessageID, "error", err)
	}
}

// clearMessageRefs drops calendar.channel / calendar.message, keeping the URL
// so a /calendar refresh can republish the embed. Only the exact refs the
// caller found gone are dropped, so a scheduler tick racing a /calendar set
// cannot wipe the refs of the message that command just published.
func clearMessageRefs(ctx context.Context, cfg guildCalendar) {
	if err := deleteMessageRefs(ctx, cfg); err != nil {
		logger.Error("Calendar: clearing message refs failed", "guild", cfg.GuildID, "error", err)
	}
}

// isGone reports whether a REST error means the target message or channel does
// not exist any more.
func isGone(err error) bool {
	var restErr *rest.Error
	if !errors.As(err, &restErr) {
		return false
	}
	return restErr.Code == rest.JSONErrorCodeUnknownMessage ||
		restErr.Code == rest.JSONErrorCodeUnknownChannel
}
