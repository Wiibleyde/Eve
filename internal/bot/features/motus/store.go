package motus

import (
	"context"
	"sync"
	"time"

	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/database/ent/activemotus"
	"Eve/internal/database/tables"

	"github.com/google/uuid"
)

// gameLocks serializes the read-modify-write cycle of a board. Motus is
// collaborative: two players can submit at the same instant and would otherwise
// both append to the same attempts snapshot, losing one guess.
var gameLocks sync.Map // messageID -> *sync.Mutex

func lockGame(messageID string) func() {
	value, _ := gameLocks.LoadOrStore(messageID, &sync.Mutex{})
	mu := value.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// forgetGameLock drops the mutex of a finished board. A late interaction that
// recreates one is harmless: it will read a terminal state and bail out.
func forgetGameLock(messageID string) {
	gameLocks.Delete(messageID)
}

func loadGame(ctx context.Context, messageID string) (*ent.ActiveMotus, error) {
	return database.Default.Ent().ActiveMotus.Query().
		Where(activemotus.MessageID(messageID)).
		Only(ctx)
}

func createGame(ctx context.Context, messageID, channelID, guildID, word, startedBy string) error {
	return database.Default.Ent().ActiveMotus.Create().
		SetID(uuid.New().String()).
		SetMessageID(messageID).
		SetChannelID(channelID).
		SetGuildID(guildID).
		SetWord(word).
		SetAttempts([]tables.MotusAttempt{}).
		SetState(tables.MotusStatePlaying).
		SetStartedBy(startedBy).
		SetExpiresAt(time.Now().Add(GameTTL)).
		Exec(ctx)
}

func saveAttempt(ctx context.Context, id string, attempts []tables.MotusAttempt, state string) error {
	return database.Default.Ent().ActiveMotus.UpdateOneID(id).
		SetAttempts(attempts).
		SetState(state).
		Exec(ctx)
}

// markState only updates the state, used when a game is closed by expiry.
func markState(ctx context.Context, id, state string) error {
	return database.Default.Ent().ActiveMotus.UpdateOneID(id).
		SetState(state).
		Exec(ctx)
}

func isExpired(game *ent.ActiveMotus) bool {
	return time.Now().After(game.ExpiresAt)
}

func isOver(game *ent.ActiveMotus) bool {
	return game.State != tables.MotusStatePlaying
}

// alreadyTried reports whether the normalized guess was already submitted.
func alreadyTried(attempts []tables.MotusAttempt, guess string) bool {
	for _, attempt := range attempts {
		if attempt.Word == guess {
			return true
		}
	}
	return false
}
