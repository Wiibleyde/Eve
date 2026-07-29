package debug

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

const RoleName = "Eve Debug"

var (
	errMissingPermissions = errors.New("debug: bot cannot manage roles")
	errStorage            = errors.New("debug: guild config access failed")
)

var guildLocks sync.Map

func lockGuild(guildID snowflake.ID) func() {
	value, _ := guildLocks.LoadOrStore(guildID, &sync.Mutex{})
	mu := value.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func ensureRole(ctx context.Context, e *events.ApplicationCommandInteractionCreate, guildID snowflake.ID) (discord.Role, error) {
	unlock := lockGuild(guildID)
	defer unlock()

	client := e.Client()
	guildKey := guildID.String()

	stored, found, err := loadRoleID(ctx, guildKey)
	if err != nil {
		return discord.Role{}, fmt.Errorf("%w: %w", errStorage, err)
	}

	if found {
		role, err := client.Rest.GetRole(guildID, stored, rest.WithCtx(ctx))
		switch {
		case err == nil:
			return *role, nil
		case isNotFound(err):
			logger.Info("Debug: stored role no longer exists, recreating",
				"guild", guildKey, "role", stored.String())
		case isMissingPermissions(err):
			return discord.Role{}, fmt.Errorf("%w: %w", errMissingPermissions, err)
		default:
			return discord.Role{}, fmt.Errorf("fetching debug role: %w", err)
		}
	}

	role, err := createRole(ctx, e, guildID)
	if err != nil {
		return discord.Role{}, err
	}

	if err := saveRoleID(ctx, guildKey, role.ID); err != nil {
		logger.Error("Debug: storing role id failed", "guild", guildKey, "role", role.ID.String(), "error", err)
		return discord.Role{}, fmt.Errorf("%w: %w", errStorage, err)
	}

	logger.Info("Debug: created managed role", "guild", guildKey, "role", role.ID.String())
	return role, nil
}

func createRole(ctx context.Context, e *events.ApplicationCommandInteractionCreate, guildID snowflake.ID) (discord.Role, error) {
	permissions := discord.PermissionsNone
	role, err := e.Client().Rest.CreateRole(guildID, discord.RoleCreate{
		Name:        RoleName,
		Permissions: &permissions,
		Hoist:       false,
		Mentionable: false,
	}, rest.WithCtx(ctx))
	if err != nil {
		if isMissingPermissions(err) {
			return discord.Role{}, fmt.Errorf("%w: %w", errMissingPermissions, err)
		}
		return discord.Role{}, fmt.Errorf("creating debug role: %w", err)
	}
	return *role, nil
}

func guildName(ctx context.Context, e *events.ApplicationCommandInteractionCreate, guildID snowflake.ID) (string, bool) {
	if guild, ok := e.Guild(); ok && guild.Name != "" {
		return guild.Name, true
	}
	guild, err := e.Client().Rest.GetGuild(guildID, false, rest.WithCtx(ctx))
	if err != nil || guild.Name == "" {
		logger.Debug("Debug: could not resolve guild name", "guild", guildID.String(), "error", err)
		return "", false
	}
	return guild.Name, true
}

func restStatus(err error) (int, bool) {
	var restErr *rest.Error
	if !errors.As(err, &restErr) || restErr.Response == nil {
		return 0, false
	}
	return restErr.Response.StatusCode, true
}

func isNotFound(err error) bool {
	status, ok := restStatus(err)
	return ok && status == http.StatusNotFound
}

func isMissingPermissions(err error) bool {
	status, ok := restStatus(err)
	return ok && status == http.StatusForbidden
}
