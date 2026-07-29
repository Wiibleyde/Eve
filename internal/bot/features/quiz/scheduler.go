package quiz

import (
	"context"
	"time"

	"Eve/internal/bot/maintenance"
	"Eve/internal/database"
	"Eve/internal/database/ent/activequiz"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
)

const (
	cleanupInterval = 24 * time.Hour
	// cleanupAfter is how long a quiz row is kept after being launched. It is
	// well past the 8h answer window, so history stays available for a while.
	cleanupAfter = 7 * 24 * time.Hour
)

// StartScheduler launches the daily cleanup of finished quizzes.
//
// Expiry itself is checked lazily by the answer handler, so this only reclaims
// rows; the client is unused but kept for a uniform scheduler signature.
func StartScheduler(_ *bot.Client) {
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		cleanup()
		for range ticker.C {
			cleanup()
		}
	}()
}

func cleanup() {
	if maintenance.Enabled() {
		logger.Debug("Quiz cleanup skipped: maintenance mode")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	n, err := database.Default.Ent().ActiveQuiz.Delete().
		Where(activequiz.LaunchedAtLT(time.Now().Add(-cleanupAfter))).
		Exec(ctx)
	if err != nil {
		logger.Error("Error cleaning up old quizzes", "error", err)
		return
	}
	if n > 0 {
		logger.Debug("Old quizzes cleaned up", "count", n)
	}
}
