package motus

import (
	"context"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
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

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	guildID := e.GuildID()
	if guildID == nil {
		helpers.RespondEphemeralCard(e, ui.Error(MsgGuildOnly))
		return
	}

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
		editResponse(client, appID, token, ui.Error(MsgNoWord))
		return
	}
	logger.Debug("Motus game starting", "channel", channelID.String(), "word", word)

	message, err := client.Rest.CreateMessage(channelID, boardMessage(word))
	if err != nil {
		logger.Error("Motus: posting the board", "error", err)
		editResponse(client, appID, token, ui.Error(MsgBoardFailed))
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
		if delErr := client.Rest.DeleteMessage(message.ChannelID, message.ID); delErr != nil {
			logger.Error("Motus: removing the orphan board", "error", delErr)
		}
		editResponse(client, appID, token, ui.Error(MsgStoreFailed))
		return
	}

	editResponse(client, appID, token, ui.Success(MsgGameStarted))
}

func editResponse(client *bot.Client, appID snowflake.ID, token string, card *ui.Card) {
	if _, err := client.Rest.UpdateInteractionResponse(appID, token, card.MessageUpdate()); err != nil {
		logger.Error("Motus: updating the interaction response", "error", err)
	}
}
