package quote

import (
	"context"
	"strings"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
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
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(MsgNoText))
		return
	}

	res, err := Create(context.Background(), e.Client(), guild, message.Author, text, "")
	if err != nil {
		logger.Error("creating quote from message", "message", message.ID.String(), "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(ErrorMessage(err)))
		return
	}
	respond(e, res)
}
