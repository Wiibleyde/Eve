package maintenance

import (
	"Eve/internal/bot/helpers"
	"Eve/internal/bot/maintenance"
	"Eve/internal/bot/ui"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const CommandName = "maintenance"

var command = discord.SlashCommandCreate{
	Name:        CommandName,
	Description: "Activer ou désactiver le mode maintenance du bot",
}

func Commands() []discord.ApplicationCommandCreate {
	if !helpers.OwnerConfigured() {
		return nil
	}
	return []discord.ApplicationCommandCreate{command}
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	if !helpers.IsOwner(e.User().ID) {
		helpers.RespondEphemeralCard(e, ui.Error("Cette commande est réservée au propriétaire du bot."))
		return
	}

	enabled := maintenance.Toggle()
	logger.Info("Maintenance mode toggled", "enabled", enabled, "by", e.User().ID.String())

	if enabled {
		helpers.RespondEphemeralCard(e, ui.Warning("Maintenance activée",
			"Mode maintenance **activé**. Les interactions des autres utilisateurs sont désormais bloquées."))
		return
	}

	helpers.RespondEphemeralCard(e, ui.Success("Mode maintenance **désactivé**. Le bot répond de nouveau normalement.").
		Title("🛠️ Maintenance désactivée"))
}
