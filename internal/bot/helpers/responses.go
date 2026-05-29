package helpers

import (
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// EphemeralResponder is satisfied by all disgo interaction event types.
type EphemeralResponder interface {
	CreateMessage(discord.MessageCreate, ...rest.RequestOpt) error
}

// RespondEphemeral sends an ephemeral message in response to an interaction.
func RespondEphemeral(r EphemeralResponder, content string) {
	if err := r.CreateMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		logger.Error("Error responding to interaction", "error", err)
	}
}

// RespondEphemeralEmbed sends an ephemeral embed in response to an interaction.
func RespondEphemeralEmbed(r EphemeralResponder, embed discord.Embed) {
	if err := r.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	}); err != nil {
		logger.Error("Error responding to interaction", "error", err)
	}
}

// RespondFollowupEphemeral sends an ephemeral follow-up message after a deferred interaction.
func RespondFollowupEphemeral(client *bot.Client, appID snowflake.ID, token string, content string) {
	if _, err := client.Rest.CreateFollowupMessage(appID, token, discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		logger.Error("Error creating followup", "error", err)
	}
}
