package music

import (
	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/events"
)

func HandleBackButton(e *events.ComponentInteractionCreate, _ []string) {
	client, ok := ready(e)
	if !ok {
		return
	}
	guildID, ok := guildOf(e, e.GuildID())
	if !ok {
		return
	}
	if _, ok := activePlayer(e, client, guildID); !ok {
		return
	}

	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("Music: deferring back button", "error", err)
		return
	}
	helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), backCard(guildID))
}

func HandleSkipButton(e *events.ComponentInteractionCreate, _ []string) {
	client, ok := ready(e)
	if !ok {
		return
	}
	guildID, ok := guildOf(e, e.GuildID())
	if !ok {
		return
	}
	if _, ok := activePlayer(e, client, guildID); !ok {
		return
	}

	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("Music: deferring skip button", "error", err)
		return
	}
	helpers.EditResponseCard(e.Client(), e.ApplicationID(), e.Token(), skipCard(guildID))
}

func HandlePlayPauseButton(e *events.ComponentInteractionCreate, _ []string) {
	client, ok := ready(e)
	if !ok {
		return
	}
	guildID, ok := guildOf(e, e.GuildID())
	if !ok {
		return
	}
	player, ok := activePlayer(e, client, guildID)
	if !ok {
		return
	}

	paused := !player.Paused
	if !setPaused(guildID, player, paused) {
		helpers.RespondEphemeralCard(e, ui.Error(MsgPlayFailed))
		return
	}

	if paused {
		helpers.RespondEphemeralCard(e, ui.Success("Musique mise en pause."))
		return
	}
	helpers.RespondEphemeralCard(e, ui.Success("Musique reprise."))
}

func HandleLoopButton(e *events.ComponentInteractionCreate, _ []string) {
	client, ok := ready(e)
	if !ok {
		return
	}
	guildID, ok := guildOf(e, e.GuildID())
	if !ok {
		return
	}
	if _, ok := activePlayer(e, client, guildID); !ok {
		return
	}

	mode := stateFor(guildID).cycleRepeat()
	refreshNowPlaying(guildID)
	helpers.RespondEphemeralCard(e, ui.Success("Boucle : **"+mode.Label()+"**"))
}
