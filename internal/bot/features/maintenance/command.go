// Package maintenance provides the owner-only /maintenance command that flips
// the global maintenance switch held by Eve/internal/bot/maintenance.
package maintenance

import (
	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/bot/maintenance"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const CommandName = "maintenance"

var command = discord.SlashCommandCreate{
	Name:        CommandName,
	Description: "Activer ou désactiver le mode maintenance du bot",
}

// Commands is a function, not a package var, because it depends on BOT_OWNER_ID
// which is only present once godotenv has run. When no owner is configured the
// command is not registered at all: nobody would be allowed to use it.
func Commands() []discord.ApplicationCommandCreate {
	if !helpers.OwnerConfigured() {
		return nil
	}
	return []discord.ApplicationCommandCreate{command}
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	if !helpers.IsOwner(e.User().ID) {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Cette commande est réservée au propriétaire du bot."))
		return
	}

	enabled := maintenance.Toggle()
	logger.Info("Maintenance mode toggled", "enabled", enabled, "by", e.User().ID.String())

	if enabled {
		embed := embeds.SuccessEmbed("Mode maintenance **activé**. Les interactions des autres utilisateurs sont désormais bloquées.")
		embed.Title = "Maintenance activée"
		helpers.RespondEphemeralEmbed(e, embed)
		return
	}

	embed := embeds.SuccessEmbed("Mode maintenance **désactivé**. Le bot répond de nouveau normalement.")
	embed.Title = "Maintenance désactivée"
	helpers.RespondEphemeralEmbed(e, embed)
}
