// Package debug exposes /debug, a self-service toggle of the managed
// «Eve Debug» marker role.
//
// The role is a pure marker: it is created by the bot with zero permissions,
// not hoisted and not mentionable, and its ID is stored in guild_configs under
// the debug.role key. Anyone may toggle it on themselves — that is harmless as
// long as the role stays permissionless, which is verified on every use.
//
// The role heals itself: a missing key, an unparsable stored ID or a role that
// was deleted from the guild all lead to a fresh role being created and stored.
package debug

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// CommandName is the slash command name, and therefore also its routing key.
const CommandName = "debug"

// commandTimeout bounds the whole handler once the interaction is acknowledged:
// one query, up to two REST calls and one upsert.
const commandTimeout = 5 * time.Second

// User-facing strings (French, per project conventions).
const (
	msgGuildOnly = "Cette commande doit être utilisée dans un serveur."
	msgDBError   = "Erreur lors de l'accès à la base de données."
	msgRoleError = "Impossible de récupérer ou de créer le rôle « " + RoleName + " »."

	msgMissingPerm = "Je n'ai pas la permission **Gérer les rôles**. " +
		"Accordez-la moi et placez mon rôle au-dessus du rôle « " + RoleName + " » pour que je puisse vous l'attribuer."

	msgEscalated = "Le rôle « " + RoleName + " » possède des permissions alors qu'il ne devrait en avoir aucune. " +
		"Par sécurité, je refuse de l'attribuer. Retirez-lui toutes ses permissions (ou supprimez-le, il sera recréé automatiquement)."
	msgEscalatedNote = "⚠️ Ce rôle avait des permissions alors qu'il ne devrait en avoir aucune : " +
		"vérifiez qui l'a modifié avant de le réutiliser."

	msgEnabledFmt      = "Vous êtes maintenant en mode debug sur le serveur %s"
	msgDisabledFmt     = "Vous n'êtes plus en mode debug sur le serveur %s"
	msgEnabledFallback = "Vous êtes maintenant en mode debug sur ce serveur"
	msgDisabledFallbck = "Vous n'êtes plus en mode debug sur ce serveur"

	msgAddFailed    = "Impossible de vous attribuer le rôle « " + RoleName + " »."
	msgRemoveFailed = "Impossible de vous retirer le rôle « " + RoleName + " »."
)

// Commands is the feature's slash command set, appended to allCommands in bot.go.
var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        CommandName,
		Description: "Activer ou désactiver le mode debug pour vous-même",
		Contexts:    []discord.InteractionContextType{discord.InteractionContextTypeGuild},
	},
}

// HandleCommand toggles the managed debug role on the invoking member.
//
// Access is deliberately open (the TS version behaved the same): the role is a
// marker with no permissions, so self-assignment grants nothing. The safety net
// is the escalation check below, not an access check.
func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	guildID := e.GuildID()
	member := e.Member()
	if guildID == nil || member == nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(msgGuildOnly))
		return
	}

	// Fail fast and explain when the bot cannot manage roles at all. When the
	// permissions are unknown (nil), the REST calls below report it instead.
	if perms := e.AppPermissions(); perms != nil && perms.Missing(discord.PermissionManageRoles) {
		logger.Debug("Debug: missing Manage Roles", "guild", guildID.String())
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(msgMissingPerm))
		return
	}

	// Resolving (and possibly creating) the role plus editing member roles is
	// up to three REST calls — acknowledge first.
	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("Debug: deferring response failed", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	role, err := ensureRole(ctx, e, *guildID)
	if err != nil {
		switch {
		case errors.Is(err, errMissingPermissions):
			logger.Warn("Debug: role management refused by Discord", "guild", guildID.String(), "error", err)
			editDeferred(e, embeds.ErrorEmbed(msgMissingPerm))
		case errors.Is(err, errStorage):
			logger.Error("Debug: guild config access failed", "guild", guildID.String(), "error", err)
			editDeferred(e, embeds.ErrorEmbed(msgDBError))
		default:
			logger.Error("Debug: resolving debug role failed", "guild", guildID.String(), "error", err)
			editDeferred(e, embeds.ErrorEmbed(msgRoleError))
		}
		return
	}

	hasRole := slices.Contains(member.RoleIDs, role.ID)

	// A managed marker role must stay permissionless. If someone escalated it,
	// never hand it out — but still let a member drop it, which can only reduce
	// their privileges.
	escalated := role.Permissions != discord.PermissionsNone
	if escalated {
		logger.Warn("Debug: managed role has permissions",
			"guild", guildID.String(),
			"role", role.ID.String(),
			"permissions", role.Permissions.String(),
		)
		if !hasRole {
			editDeferred(e, embeds.ErrorEmbed(msgEscalated))
			return
		}
	}

	if hasRole {
		removeRole(ctx, e, *guildID, role.ID, escalated)
		return
	}
	addRole(ctx, e, *guildID, role.ID)
}

func addRole(ctx context.Context, e *events.ApplicationCommandInteractionCreate, guildID snowflake.ID, roleID snowflake.ID) {
	userID := e.User().ID
	if err := e.Client().Rest.AddMemberRole(guildID, userID, roleID, rest.WithCtx(ctx)); err != nil {
		logger.Error("Debug: adding role failed", "guild", guildID.String(), "user", userID.String(), "error", err)
		editDeferred(e, embeds.ErrorEmbed(toggleErrorMessage(err, msgAddFailed)))
		return
	}
	logger.Info("Debug: mode enabled", "guild", guildID.String(), "user", userID.String())
	editDeferred(e, embeds.SuccessEmbed(toggleMessage(ctx, e, guildID, msgEnabledFmt, msgEnabledFallback)))
}

// removeRole revokes the debug role from the invoking user. When the role had
// been escalated, the reply also carries the warning.
func removeRole(ctx context.Context, e *events.ApplicationCommandInteractionCreate, guildID snowflake.ID, roleID snowflake.ID, escalated bool) {
	userID := e.User().ID
	if err := e.Client().Rest.RemoveMemberRole(guildID, userID, roleID, rest.WithCtx(ctx)); err != nil {
		logger.Error("Debug: removing role failed", "guild", guildID.String(), "user", userID.String(), "error", err)
		editDeferred(e, embeds.ErrorEmbed(toggleErrorMessage(err, msgRemoveFailed)))
		return
	}
	logger.Info("Debug: mode disabled", "guild", guildID.String(), "user", userID.String())

	message := toggleMessage(ctx, e, guildID, msgDisabledFmt, msgDisabledFallbck)
	if escalated {
		message += "\n\n" + msgEscalatedNote
	}
	editDeferred(e, embeds.SuccessEmbed(message))
}

// toggleMessage renders the toggle confirmation, naming the guild when its name
// is known and falling back to a name-free wording otherwise.
func toggleMessage(ctx context.Context, e *events.ApplicationCommandInteractionCreate, guildID snowflake.ID, format string, fallback string) string {
	name, ok := guildName(ctx, e, guildID)
	if !ok {
		return fallback
	}
	return fmt.Sprintf(format, name)
}

// toggleErrorMessage turns a role edit failure into a user-facing message: a
// 403 always means either the missing permission or an unreachable role in the
// hierarchy, both covered by msgMissingPerm.
func toggleErrorMessage(err error, fallback string) string {
	if isMissingPermissions(err) {
		return msgMissingPerm
	}
	return fallback
}

// editDeferred replaces the deferred ephemeral response with an embed.
func editDeferred(e *events.ApplicationCommandInteractionCreate, embed discord.Embed) {
	embedList := []discord.Embed{embed}
	if _, err := e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.MessageUpdate{
		Embeds: &embedList,
	}); err != nil {
		logger.Error("Debug: editing deferred response failed", "error", err)
	}
}
