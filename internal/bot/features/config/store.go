package config

import (
	"context"

	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/guildconfig"
)

// Value returns the raw stored value of a configuration key for a guild.
// The second result reports whether the key is set.
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

// BoolValue returns a Bool-kind key value, falling back to def when unset.
func BoolValue(ctx context.Context, guildID string, key string, def bool) (bool, error) {
	raw, ok, err := Value(ctx, guildID, key)
	if err != nil || !ok {
		return def, err
	}
	return parseBool(raw), nil
}

// setValue upserts a configuration entry. ent has no OnConflict support in this
// project (the sql/upsert feature is off), so it updates first and inserts only
// when no row matched — same pattern as the rest of the codebase.
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

// visibleValues returns the stored values of every user-visible key of a guild,
// indexed by key name. Keys that are not set are simply absent.
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
