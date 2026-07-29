package helpers

import (
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
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		logger.Error("Error responding to interaction", "error", err)
	}
}

func RespondEphemeralEmbed(r EphemeralResponder, embed discord.Embed) {
	if err := r.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	}); err != nil {
		logger.Error("Error responding to interaction", "error", err)
	}
}

func RespondFollowupEphemeral(client *bot.Client, appID snowflake.ID, token string, content string) {
	if _, err := client.Rest.CreateFollowupMessage(appID, token, discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		logger.Error("Error creating followup", "error", err)
	}
}
