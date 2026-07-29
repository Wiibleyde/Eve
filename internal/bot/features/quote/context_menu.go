package quote

import (
	"context"
	"strings"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/events"
)

const MenuCreateQuote = "Créer une citation"

func HandleCreateQuoteMenu(e *events.ApplicationCommandInteractionCreate) {
	guild, ok := guildID(e)
	if !ok {
		return
	}

	message := e.MessageCommandInteractionData().TargetMessage()
	text := strings.TrimSpace(message.Content)
	if text == "" {
		helpers.RespondEphemeralCard(e, ui.Error(MsgNoText))
		return
	}

	res, err := Create(context.Background(), e.Client(), guild, message.Author, text, "")
	if err != nil {
		logger.Error("creating quote from message", "message", message.ID.String(), "error", err)
		helpers.RespondEphemeralCard(e, ui.Error(ErrorMessage(err)))
		return
	}
	respond(e, res)
}
