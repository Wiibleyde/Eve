package debug

import (
	"context"
	"fmt"
	"strings"

	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/guildconfig"
	"Eve/internal/database/tables"
	"Eve/internal/logger"

	"github.com/disgoorg/snowflake/v2"
)

// loadRoleID reads the debug.role key of a guild.
//
// The bool reports whether a usable role ID is stored: a missing row and a
// corrupted value are both reported as "not stored", so the caller simply
// creates a fresh role instead of erroring out.
func loadRoleID(ctx context.Context, guildID string) (snowflake.ID, bool, error) {
	row, err := database.Default.Ent().GuildConfig.Query().
		Where(guildconfig.GuildID(guildID), guildconfig.Key(tables.DebugRole.String())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("querying debug role: %w", err)
	}

	roleID, parseErr := snowflake.Parse(strings.TrimSpace(row.Value))
	if parseErr != nil {
		logger.Warn("Debug: stored role id is not a snowflake, recreating the role",
			"guild", guildID, "value", row.Value)
		return 0, false, nil
	}
	return roleID, true, nil
}

// saveRoleID upserts the debug.role key, mirroring the pattern used by the
// config feature (update first, insert when nothing was updated) since the
// sql/upsert ent feature is off in this project.
func saveRoleID(ctx context.Context, guildID string, roleID snowflake.ID) error {
	name := tables.DebugRole.String()

	updated, err := database.Default.Ent().GuildConfig.Update().
		Where(guildconfig.GuildID(guildID), guildconfig.Key(name)).
		SetValue(roleID.String()).
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
		SetValue(roleID.String()).
		Exec(ctx); err != nil {
		return fmt.Errorf("creating %s: %w", name, err)
	}
	return nil
}
