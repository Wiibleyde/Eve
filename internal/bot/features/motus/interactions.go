package motus

import (
	"context"
	"fmt"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/bot/router"
	"Eve/internal/database/ent"
	"Eve/internal/database/tables"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

// HandleTryButton opens the guess modal for the board the button belongs to.
func HandleTryButton(e *events.ComponentInteractionCreate, _ []string) {
	messageID := e.Message.ID
	ctx := context.Background()

	game, ok := fetchGame(ctx, e, messageID.String())
	if !ok {
		return
	}

	if isOver(game) {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(MsgGameOver, game.Word)))
		return
	}
	if isExpired(game) {
		closeExpiredGame(ctx, e.Client(), game)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(MsgExpired, game.Word)))
		return
	}

	// Max is the word length: accented input still fits since precomposed
	// accents count one character. Min stays 1 because ligatures type shorter
	// than they normalize («cœur» is 4 characters for 5 letters); the exact
	// length is re-validated server side either way.
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

// HandleSubmitModal validates a guess, updates the game and edits the board.
// Invalid guesses never consume an attempt.
func HandleSubmitModal(e *events.ModalSubmitInteractionCreate, args []string) {
	messageID := ""
	if len(args) > 0 {
		messageID = args[0]
	}
	if messageID == "" && e.Message != nil {
		messageID = e.Message.ID.String()
	}
	if messageID == "" {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(MsgUnknownGame))
		return
	}

	ctx := context.Background()

	// Motus is collaborative: serialize the read-modify-write of one board.
	unlock := lockGame(messageID)
	defer unlock()

	game, ok := fetchGame(ctx, e, messageID)
	if !ok {
		return
	}

	if isOver(game) {
		// The game ended through another path (win/lose/expiry) that already
		// dropped its lock entry, but this call itself re-created one via
		// lockGame above: drop it again so a late resubmission never pins a
		// stale mutex in gameLocks forever.
		forgetGameLock(messageID)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(MsgGameOver, game.Word)))
		return
	}
	if isExpired(game) {
		closeExpiredGame(ctx, e.Client(), game)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(MsgExpired, game.Word)))
		return
	}

	guess := Normalize(e.Data.Text(InputGuess))
	expected := WordLen(game.Word)
	if !IsWord(guess) || WordLen(guess) != expected {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(MsgInvalidGuess, expected)))
		return
	}
	if alreadyTried(game.Attempts, guess) {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(MsgAlreadyTried, guess)))
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
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(MsgDBError))
		return
	}

	editBoard(e.Client(), game, attempts, state)

	switch state {
	case tables.MotusStateWon:
		forgetGameLock(messageID)
		helpers.RespondEphemeralEmbed(e, embeds.SuccessEmbed(
			fmt.Sprintf(MsgWon, game.Word, len(attempts), MaxAttempts)))
	case tables.MotusStateLost:
		forgetGameLock(messageID)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(MsgLost, game.Word)))
	default:
		helpers.RespondEphemeralEmbed(e, attemptFeedback(game.Word, attempts))
	}
}

// fetchGame loads a game and answers the interaction itself when it cannot.
func fetchGame(ctx context.Context, responder helpers.EphemeralResponder, messageID string) (*ent.ActiveMotus, bool) {
	game, err := loadGame(ctx, messageID)
	if err == nil {
		return game, true
	}
	if ent.IsNotFound(err) {
		helpers.RespondEphemeralEmbed(responder, embeds.ErrorEmbed(MsgUnknownGame))
		return nil, false
	}
	logger.Error("Motus: loading the game", "message", messageID, "error", err)
	helpers.RespondEphemeralEmbed(responder, embeds.ErrorEmbed(MsgDBError))
	return nil, false
}

// closeExpiredGame marks an outlived game as lost and reveals the answer on the
// public board. Expiry is lazy: it only happens when someone touches the game.
// Callers must have already checked isOver: closing a game that ended by play
// (won/lost) would overwrite its board with a false "Perdu !".
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
