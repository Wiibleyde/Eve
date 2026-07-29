package motus

import (
	"context"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        CommandName,
		Description: "Lancer une partie de Motus : un mot à deviner en 6 essais",
		Contexts:    []discord.InteractionContextType{discord.InteractionContextTypeGuild},
	},
}

// HandleCommand starts a game: it posts the public board with its «Essayer»
// button and answers the launcher ephemerally.
func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	guildID := e.GuildID()
	if guildID == nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(MsgGuildOnly))
		return
	}

	// Picking a word hits the network, which can outlive the 3s interaction
	// window: acknowledge first, then edit the deferred response.
	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("Motus: deferring the command response", "error", err)
		return
	}

	client := e.Client()
	appID := e.ApplicationID()
	token := e.Token()
	channelID := e.Channel().ID()
	ctx := context.Background()

	word := PickWord(ctx)
	if word == "" {
		editResponse(client, appID, token, embeds.ErrorEmbed(MsgNoWord))
		return
	}
	// DEBUG only: the answer must never reach INFO and above.
	logger.Debug("Motus game starting", "channel", channelID.String(), "word", word)

	message, err := client.Rest.CreateMessage(channelID, boardMessage(word))
	if err != nil {
		logger.Error("Motus: posting the board", "error", err)
		editResponse(client, appID, token, embeds.ErrorEmbed(MsgBoardFailed))
		return
	}

	err = createGame(ctx,
		message.ID.String(),
		message.ChannelID.String(),
		guildID.String(),
		word,
		e.User().ID.String(),
	)
	if err != nil {
		logger.Error("Motus: storing the game", "error", err)
		// An unplayable board is worse than no board at all.
		if delErr := client.Rest.DeleteMessage(message.ChannelID, message.ID); delErr != nil {
			logger.Error("Motus: removing the orphan board", "error", delErr)
		}
		editResponse(client, appID, token, embeds.ErrorEmbed(MsgStoreFailed))
		return
	}

	editResponse(client, appID, token, embeds.SuccessEmbed(MsgGameStarted))
}

// editResponse replaces the deferred ephemeral response with an embed.
func editResponse(client *bot.Client, appID snowflake.ID, token string, embed discord.Embed) {
	embedList := []discord.Embed{embed}
	if _, err := client.Rest.UpdateInteractionResponse(appID, token, discord.MessageUpdate{
		Embeds: &embedList,
	}); err != nil {
		logger.Error("Motus: updating the interaction response", "error", err)
	}
}
