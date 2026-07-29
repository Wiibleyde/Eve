package motus

import (
	"context"
	"fmt"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/router"
	"Eve/internal/bot/ui"
	"Eve/internal/database/ent"
	"Eve/internal/database/tables"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func HandleTryButton(e *events.ComponentInteractionCreate, _ []string) {
	messageID := e.Message.ID
	ctx := context.Background()

	game, ok := fetchGame(ctx, e, messageID.String())
	if !ok {
		return
	}

	if isOver(game) {
		helpers.RespondEphemeralCard(e, ui.Error(fmt.Sprintf(MsgGameOver, game.Word)))
		return
	}
	if isExpired(game) {
		closeExpiredGame(ctx, e.Client(), game)
		helpers.RespondEphemeralCard(e, ui.Error(fmt.Sprintf(MsgExpired, game.Word)))
		return
	}

	length := WordLen(game.Word)
	input := discord.NewShortTextInput(InputGuess).
		WithRequired(true).
		WithMinLength(1).
		WithMaxLength(length).
		WithPlaceholder(fmt.Sprintf("%d lettres, commence par %s", length, FirstLetter(game.Word)))

	modal := discord.NewModalCreate(
		router.BuildCustomID(CustomIDSubmit, messageID.String()),
		"Motus — votre proposition",
		discord.NewLabel(fmt.Sprintf("Mot de %d lettres", length), input),
	)

	if err := e.Modal(modal); err != nil {
		logger.Error("Motus: opening the guess modal", "error", err)
	}
}

func HandleSubmitModal(e *events.ModalSubmitInteractionCreate, args []string) {
	messageID := ""
	if len(args) > 0 {
		messageID = args[0]
	}
	if messageID == "" && e.Message != nil {
		messageID = e.Message.ID.String()
	}
	if messageID == "" {
		helpers.RespondEphemeralCard(e, ui.Error(MsgUnknownGame))
		return
	}

	ctx := context.Background()

	unlock := lockGame(messageID)
	defer unlock()

	game, ok := fetchGame(ctx, e, messageID)
	if !ok {
		return
	}

	if isOver(game) {
		forgetGameLock(messageID)
		helpers.RespondEphemeralCard(e, ui.Error(fmt.Sprintf(MsgGameOver, game.Word)))
		return
	}
	if isExpired(game) {
		closeExpiredGame(ctx, e.Client(), game)
		helpers.RespondEphemeralCard(e, ui.Error(fmt.Sprintf(MsgExpired, game.Word)))
		return
	}

	guess := Normalize(e.Data.Text(InputGuess))
	expected := WordLen(game.Word)
	if !IsWord(guess) || WordLen(guess) != expected {
		helpers.RespondEphemeralCard(e, ui.Error(fmt.Sprintf(MsgInvalidGuess, expected)))
		return
	}
	if alreadyTried(game.Attempts, guess) {
		helpers.RespondEphemeralCard(e, ui.Error(fmt.Sprintf(MsgAlreadyTried, guess)))
		return
	}

	attempts := make([]tables.MotusAttempt, 0, len(game.Attempts)+1)
	attempts = append(attempts, game.Attempts...)
	attempts = append(attempts, tables.MotusAttempt{Word: guess, UserID: e.User().ID.String()})

	state := tables.MotusStatePlaying
	switch {
	case IsWinning(Score(guess, game.Word)):
		state = tables.MotusStateWon
	case len(attempts) >= MaxAttempts:
		state = tables.MotusStateLost
	}

	if err := saveAttempt(ctx, game.ID, attempts, state); err != nil {
		logger.Error("Motus: saving the attempt", "error", err)
		helpers.RespondEphemeralCard(e, ui.Error(MsgDBError))
		return
	}

	editBoard(e.Client(), game, attempts, state)

	switch state {
	case tables.MotusStateWon:
		forgetGameLock(messageID)
		helpers.RespondEphemeralCard(e, ui.Success(
			fmt.Sprintf(MsgWon, game.Word, len(attempts), MaxAttempts)))
	case tables.MotusStateLost:
		forgetGameLock(messageID)
		helpers.RespondEphemeralCard(e, ui.Error(fmt.Sprintf(MsgLost, game.Word)))
	default:
		helpers.RespondEphemeralCard(e, attemptFeedback(game.Word, attempts))
	}
}

func fetchGame(ctx context.Context, responder helpers.EphemeralResponder, messageID string) (*ent.ActiveMotus, bool) {
	game, err := loadGame(ctx, messageID)
	if err == nil {
		return game, true
	}
	if ent.IsNotFound(err) {
		helpers.RespondEphemeralCard(responder, ui.Error(MsgUnknownGame))
		return nil, false
	}
	logger.Error("Motus: loading the game", "message", messageID, "error", err)
	helpers.RespondEphemeralCard(responder, ui.Error(MsgDBError))
	return nil, false
}

func closeExpiredGame(ctx context.Context, client *bot.Client, game *ent.ActiveMotus) {
	if isOver(game) {
		return
	}
	if err := markState(ctx, game.ID, tables.MotusStateLost); err != nil {
		logger.Error("Motus: closing the expired game", "error", err)
	}
	editBoard(client, game, game.Attempts, tables.MotusStateLost)
	forgetGameLock(game.MessageID)
}

func editBoard(client *bot.Client, game *ent.ActiveMotus, attempts []tables.MotusAttempt, state string) {
	channelID, err := snowflake.Parse(game.ChannelID)
	if err != nil {
		logger.Error("Motus: parsing the channel ID", "value", game.ChannelID, "error", err)
		return
	}
	messageID, err := snowflake.Parse(game.MessageID)
	if err != nil {
		logger.Error("Motus: parsing the message ID", "value", game.MessageID, "error", err)
		return
	}

	if _, err := client.Rest.UpdateMessage(channelID, messageID, boardUpdate(game.Word, attempts, state)); err != nil {
		logger.Error("Motus: updating the board", "error", err)
	}
}
