package debug

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

const CommandName = "debug"

const commandTimeout = 5 * time.Second

const (
	msgGuildOnly = "Cette commande doit être utilisée dans un serveur."
	msgDBError   = "Erreur lors de l'accès à la base de données."
	msgRoleError = "Impossible de récupérer ou de créer le rôle « " + RoleName + " »."

	msgMissingPerm = "Je n'ai pas la permission **Gérer les rôles**. " +
		"Accordez-la moi et placez mon rôle au-dessus du rôle « " + RoleName + " » pour que je puisse vous l'attribuer."

	msgEscalated = "Le rôle « " + RoleName + " » possède la permission **Administrateur** alors qu'il ne devrait avoir aucune permission. " +
		"Par sécurité, je refuse de l'attribuer. Retirez-lui toutes ses permissions (ou supprimez-le, il sera recréé automatiquement)."
	msgEscalatedNote = "⚠️ Ce rôle avait la permission **Administrateur** alors qu'il ne devrait avoir aucune permission : " +
		"vérifiez qui l'a modifié avant de le réutiliser."

	msgEnabledFmt      = "Vous êtes maintenant en mode debug sur le serveur %s"
	msgDisabledFmt     = "Vous n'êtes plus en mode debug sur le serveur %s"
	msgEnabledFallback = "Vous êtes maintenant en mode debug sur ce serveur"
	msgDisabledFallbck = "Vous n'êtes plus en mode debug sur ce serveur"

	msgAddFailed    = "Impossible de vous attribuer le rôle « " + RoleName + " »."
	msgRemoveFailed = "Impossible de vous retirer le rôle « " + RoleName + " »."
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        CommandName,
		Description: "Activer ou désactiver le mode debug pour vous-même",
		Contexts:    []discord.InteractionContextType{discord.InteractionContextTypeGuild},
	},
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	guildID := e.GuildID()
	member := e.Member()
	if guildID == nil || member == nil {
		helpers.RespondEphemeralCard(e, ui.Error(msgGuildOnly))
		return
	}

	if perms := e.AppPermissions(); perms != nil && perms.Missing(discord.PermissionManageRoles) {
		logger.Debug("Debug: missing Manage Roles", "guild", guildID.String())
		helpers.RespondEphemeralCard(e, ui.Error(msgMissingPerm))
		return
	}

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
			editDeferred(e, ui.Error(msgMissingPerm))
		case errors.Is(err, errStorage):
			logger.Error("Debug: guild config access failed", "guild", guildID.String(), "error", err)
			editDeferred(e, ui.Error(msgDBError))
		default:
			logger.Error("Debug: resolving debug role failed", "guild", guildID.String(), "error", err)
			editDeferred(e, ui.Error(msgRoleError))
		}
		return
	}

	hasRole := slices.Contains(member.RoleIDs, role.ID)

	escalated := role.Permissions.Has(discord.PermissionAdministrator)
	if escalated {
		logger.Warn("Debug: managed role has administrator permission",
			"guild", guildID.String(),
			"role", role.ID.String(),
			"permissions", role.Permissions.String(),
		)
		if !hasRole {
			editDeferred(e, ui.Error(msgEscalated))
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
		editDeferred(e, ui.Error(toggleErrorMessage(err, msgAddFailed)))
		return
	}
	logger.Info("Debug: mode enabled", "guild", guildID.String(), "user", userID.String())
	editDeferred(e, ui.Success(toggleMessage(ctx, e, guildID, msgEnabledFmt, msgEnabledFallback)))
}

func removeRole(ctx context.Context, e *events.ApplicationCommandInteractionCreate, guildID snowflake.ID, roleID snowflake.ID, escalated bool) {
	userID := e.User().ID
	if err := e.Client().Rest.RemoveMemberRole(guildID, userID, roleID, rest.WithCtx(ctx)); err != nil {
		logger.Error("Debug: removing role failed", "guild", guildID.String(), "user", userID.String(), "error", err)
		editDeferred(e, ui.Error(toggleErrorMessage(err, msgRemoveFailed)))
		return
	}
	logger.Info("Debug: mode disabled", "guild", guildID.String(), "user", userID.String())

	message := toggleMessage(ctx, e, guildID, msgDisabledFmt, msgDisabledFallbck)
	if escalated {
		message += "\n\n" + msgEscalatedNote
	}
	editDeferred(e, ui.Success(message))
}

func toggleMessage(ctx context.Context, e *events.ApplicationCommandInteractionCreate, guildID snowflake.ID, format string, fallback string) string {
	name, ok := guildName(ctx, e, guildID)
	if !ok {
		return fallback
	}
	return fmt.Sprintf(format, name)
}

func toggleErrorMessage(err error, fallback string) string {
	if isMissingPermissions(err) {
		return msgMissingPerm
	}
	return fallback
}

func editDeferred(e *events.ApplicationCommandInteractionCreate, card *ui.Card) {
	helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), card)
}
