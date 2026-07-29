package config

import (
	"context"

	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/guildconfig"
)

func Value(ctx context.Context, guildID string, key string) (string, bool, error) {
	cfg, err := database.Default.Ent().GuildConfig.Query().
		Where(guildconfig.GuildID(guildID), guildconfig.Key(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return cfg.Value, true, nil
}

func BoolValue(ctx context.Context, guildID string, key string, def bool) (bool, error) {
	raw, ok, err := Value(ctx, guildID, key)
	if err != nil || !ok {
		return def, err
	}
	return parseBool(raw), nil
}

func setValue(ctx context.Context, guildID string, key string, value string) error {
	n, err := database.Default.Ent().GuildConfig.Update().
		Where(guildconfig.GuildID(guildID), guildconfig.Key(key)).
		SetValue(value).
		Save(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return database.Default.Ent().GuildConfig.Create().
		SetGuildID(guildID).
		SetKey(key).
		SetValue(value).
		Exec(ctx)
}

func resetValue(ctx context.Context, guildID string, key string) (int, error) {
	return database.Default.Ent().GuildConfig.Delete().
		Where(guildconfig.GuildID(guildID), guildconfig.Key(key)).
		Exec(ctx)
}

func visibleValues(ctx context.Context, guildID string) (map[string]string, error) {
	cfgs, err := database.Default.Ent().GuildConfig.Query().
		Where(guildconfig.GuildID(guildID), guildconfig.KeyIn(keyNames()...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	values := make(map[string]string, len(cfgs))
	for _, cfg := range cfgs {
		values[cfg.Key] = cfg.Value
	}
	return values, nil
}
