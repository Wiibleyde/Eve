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

// RoleName is the name of the managed marker role created per guild.
const RoleName = "Eve Debug"

// Sentinel errors, used to map a failure to the right user-facing message.
var (
	errMissingPermissions = errors.New("debug: bot cannot manage roles")
	errStorage            = errors.New("debug: guild config access failed")
)

// guildLocks serialises /debug per guild: two simultaneous invocations in a
// guild without a stored role would otherwise create two «Eve Debug» roles.
var guildLocks sync.Map // snowflake.ID -> *sync.Mutex

func lockGuild(guildID snowflake.ID) func() {
	value, _ := guildLocks.LoadOrStore(guildID, &sync.Mutex{})
	mu := value.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// ensureRole returns the guild's managed debug role.
//
// It creates and stores a fresh role when guild_configs has no usable ID or
// when the stored role no longer exists in the guild — which is what makes a
// manual deletion self-healing.
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
		// Read the role from the API rather than the cache: the permission
		// check below is a security check and must see the current state.
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
		// The role exists but could not be remembered: deleting it again would
		// be worse (the next call would recreate one anyway), so report the
		// storage failure and leave the role in place for the retry to reuse.
		logger.Error("Debug: storing role id failed", "guild", guildKey, "role", role.ID.String(), "error", err)
		return discord.Role{}, fmt.Errorf("%w: %w", errStorage, err)
	}

	logger.Info("Debug: created managed role", "guild", guildKey, "role", role.ID.String())
	return role, nil
}

// createRole creates the marker role: zero permissions, not hoisted, not
// mentionable. Permissions are sent explicitly because Discord otherwise copies
// the @everyone permissions onto the new role.
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

// guildName resolves the guild name for the confirmation message, preferring
// the cache and falling back to one REST call. The bool reports success: the
// caller then uses a wording that does not name the guild.
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

// isNotFound reports whether the API answered 404 (role deleted).
func isNotFound(err error) bool {
	status, ok := restStatus(err)
	return ok && status == http.StatusNotFound
}

// isMissingPermissions reports whether the API answered 403, which covers both
// the missing Manage Roles permission and a target role above the bot's
// highest role.
func isMissingPermissions(err error) bool {
	status, ok := restStatus(err)
	return ok && status == http.StatusForbidden
}
