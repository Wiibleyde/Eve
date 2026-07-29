package quote

import (
	"context"
	"strings"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/events"
)

// MenuCreateQuote is the name of the message context menu entry, and therefore
// also its routing key.
const MenuCreateQuote = "Créer une citation"

// HandleCreateQuoteMenu quotes the targeted message: its author is the quote
// author, its content the quote, and there is no context to attach.
func HandleCreateQuoteMenu(e *events.ApplicationCommandInteractionCreate) {
	guild, ok := guildID(e)
	if !ok {
		return
	}

	message := e.MessageCommandInteractionData().TargetMessage()
	text := strings.TrimSpace(message.Content)
	if text == "" {
		// Embed-only, attachment-only or sticker-only message: nothing to quote.
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
