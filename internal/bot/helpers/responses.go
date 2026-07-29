package helpers

import (
	"Eve/internal/bot/ui"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

type EphemeralResponder interface {
	CreateMessage(discord.MessageCreate, ...rest.RequestOpt) error
}

func RespondEphemeral(r EphemeralResponder, content string) {
	if err := r.CreateMessage(discord.MessageCreate{
		Content:         content,
		Flags:           discord.MessageFlagEphemeral,
		AllowedMentions: ui.NoMentions(),
	}); err != nil {
		logger.Error("Error responding to interaction", "error", err)
	}
}

func RespondEphemeralCard(r EphemeralResponder, card *ui.Card) {
	if err := r.CreateMessage(card.EphemeralCreate()); err != nil {
		logger.Error("Error responding to interaction", "error", err)
	}
}

func RespondCard(r EphemeralResponder, card *ui.Card) {
	if err := r.CreateMessage(card.MessageCreate()); err != nil {
		logger.Error("Error responding to interaction", "error", err)
	}
}

func RespondFollowupEphemeral(client *bot.Client, appID snowflake.ID, token string, content string) {
	if _, err := client.Rest.CreateFollowupMessage(appID, token, discord.MessageCreate{
		Content:         content,
		Flags:           discord.MessageFlagEphemeral,
		AllowedMentions: ui.NoMentions(),
	}); err != nil {
		logger.Error("Error creating followup", "error", err)
	}
}

func FollowupEphemeralCard(client *bot.Client, appID snowflake.ID, token string, card *ui.Card) {
	if _, err := client.Rest.CreateFollowupMessage(appID, token, card.EphemeralCreate()); err != nil {
		logger.Error("Error creating followup", "error", err)
	}
}

func EditResponseCard(client *bot.Client, appID snowflake.ID, token string, card *ui.Card) {
	if _, err := client.Rest.UpdateInteractionResponse(appID, token, card.MessageUpdate()); err != nil {
		logger.Error("Error editing deferred response", "error", err)
	}
}
