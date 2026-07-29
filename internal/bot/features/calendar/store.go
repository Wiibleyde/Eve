package calendar

import (
	"context"
	"fmt"

	"Eve/internal/database"
	"Eve/internal/database/ent/guildconfig"
	"Eve/internal/database/tables"
)

type guildCalendar struct {
	GuildID   string
	URL       string
	ChannelID string
	MessageID string
}

func (c guildCalendar) configured() bool { return c.URL != "" }

func (c guildCalendar) hasMessage() bool { return c.ChannelID != "" && c.MessageID != "" }

func loadGuildCalendar(ctx context.Context, guildID string) (guildCalendar, error) {
	rows, err := database.Default.Ent().GuildConfig.Query().
		Where(
			guildconfig.GuildID(guildID),
			guildconfig.KeyIn(
				tables.CalendarURL.String(),
				tables.CalendarChannel.String(),
				tables.CalendarMessage.String(),
			),
		).
		All(ctx)
	if err != nil {
		return guildCalendar{}, fmt.Errorf("querying calendar config: %w", err)
	}

	cfg := guildCalendar{GuildID: guildID}
	for _, row := range rows {
		switch row.Key {
		case tables.CalendarURL.String():
			cfg.URL = row.Value
		case tables.CalendarChannel.String():
			cfg.ChannelID = row.Value
		case tables.CalendarMessage.String():
			cfg.MessageID = row.Value
		}
	}
	return cfg, nil
}

func calendarURLs(ctx context.Context) (map[string]string, error) {
	rows, err := database.Default.Ent().GuildConfig.Query().
		Where(guildconfig.Key(tables.CalendarURL.String())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying calendar urls: %w", err)
	}

	urls := make(map[string]string, len(rows))
	for _, row := range rows {
		if row.Value == "" {
			continue
		}
		urls[row.GuildID] = row.Value
	}
	return urls, nil
}

func saveConfig(ctx context.Context, guildID string, key tables.ConfigKey, value string) error {
	name := key.String()

	updated, err := database.Default.Ent().GuildConfig.Update().
		Where(guildconfig.GuildID(guildID), guildconfig.Key(name)).
		SetValue(value).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("updating %s: %w", name, err)
	}
	if updated > 0 {
		return nil
	}

	if err := database.Default.Ent().GuildConfig.Create().
		SetGuildID(guildID).
		SetKey(name).
		SetValue(value).
		Exec(ctx); err != nil {
		return fmt.Errorf("creating %s: %w", name, err)
	}
	return nil
}

func deleteMessageRefs(ctx context.Context, cfg guildCalendar) error {
	if !cfg.hasMessage() {
		return nil
	}

	if _, err := database.Default.Ent().GuildConfig.Delete().
		Where(
			guildconfig.GuildID(cfg.GuildID),
			guildconfig.Or(
				guildconfig.And(
					guildconfig.Key(tables.CalendarChannel.String()),
					guildconfig.Value(cfg.ChannelID),
				),
				guildconfig.And(
					guildconfig.Key(tables.CalendarMessage.String()),
					guildconfig.Value(cfg.MessageID),
				),
			),
		).
		Exec(ctx); err != nil {
		return fmt.Errorf("deleting calendar message refs: %w", err)
	}
	return nil
}

func deleteConfig(ctx context.Context, guildID string, keys ...tables.ConfigKey) error {
	names := make([]string, 0, len(keys))
	for _, key := range keys {
		names = append(names, key.String())
	}
	if len(names) == 0 {
		return nil
	}

	if _, err := database.Default.Ent().GuildConfig.Delete().
		Where(guildconfig.GuildID(guildID), guildconfig.KeyIn(names...)).
		Exec(ctx); err != nil {
		return fmt.Errorf("deleting calendar config: %w", err)
	}
	return nil
}
