package calendar

import (
	"context"
	"errors"
	"fmt"

	"Eve/internal/bot/ui"
	"Eve/internal/database/tables"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

var errCalendarMessageGone = errors.New("calendar message no longer exists")

var errNoCalendarMessage = errors.New("no calendar message stored")

func updateCalendarMessage(ctx context.Context, client *bot.Client, cfg guildCalendar, card *ui.Card) error {
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

	if _, err := client.Rest.UpdateMessage(channelID, messageID, card.MessageUpdate()); err != nil {
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

func postCalendarMessage(ctx context.Context, client *bot.Client, guildID string, channelID snowflake.ID, card *ui.Card) error {
	message, err := client.Rest.CreateMessage(channelID, card.MessageCreate())
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

func clearMessageRefs(ctx context.Context, cfg guildCalendar) {
	if err := deleteMessageRefs(ctx, cfg); err != nil {
		logger.Error("Calendar: clearing message refs failed", "guild", cfg.GuildID, "error", err)
	}
}

func isGone(err error) bool {
	var restErr *rest.Error
	if !errors.As(err, &restErr) {
		return false
	}
	return restErr.Code == rest.JSONErrorCodeUnknownMessage ||
		restErr.Code == rest.JSONErrorCodeUnknownChannel
}
