package mpthreads

import (
	"context"
	"sync"

	"Eve/internal/database"
	"Eve/internal/database/ent/mpthread"
	"Eve/internal/logger"

	"github.com/disgoorg/snowflake/v2"
)

// The DB is the source of truth; these maps are the hot path used by the
// MessageCreate listener so that every DM does not hit Postgres.
var (
	cacheMu      sync.RWMutex
	threadToUser = make(map[snowflake.ID]snowflake.ID)
	userToThread = make(map[snowflake.ID]snowflake.ID)
)

// The lock only has to serialise a conversation with itself, so a fixed pool of
// mutexes shared between unrelated users is enough — and, unlike a per-user
// mutex map, it does not grow with every user who ever DMs the bot.
const userLockShards = 64

var userLocks [userLockShards]sync.Mutex

// Init loads every persisted mapping into memory. Rows with unparsable
// snowflakes are dropped from the cache (never from the DB) and logged.
func Init(ctx context.Context) error {
	rows, err := database.Default.Ent().MPThread.Query().All(ctx)
	if err != nil {
		return err
	}

	threads := make(map[snowflake.ID]snowflake.ID, len(rows))
	users := make(map[snowflake.ID]snowflake.ID, len(rows))
	for _, row := range rows {
		userID, err := snowflake.Parse(row.UserID)
		if err != nil {
			logger.Warn("MP threads: ignoring row with invalid user id", "user", row.UserID, "error", err)
			continue
		}
		threadID, err := snowflake.Parse(row.ThreadID)
		if err != nil {
			logger.Warn("MP threads: ignoring row with invalid thread id", "thread", row.ThreadID, "error", err)
			continue
		}
		threads[threadID] = userID
		users[userID] = threadID
	}

	cacheMu.Lock()
	threadToUser = threads
	userToThread = users
	cacheMu.Unlock()
	return nil
}

// MappingCount returns how many user↔thread mappings are currently cached.
func MappingCount() int {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return len(threadToUser)
}

func userForThread(threadID snowflake.ID) (snowflake.ID, bool) {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	userID, ok := threadToUser[threadID]
	return userID, ok
}

func threadForUser(userID snowflake.ID) (snowflake.ID, bool) {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	threadID, ok := userToThread[userID]
	return threadID, ok
}

// rememberMapping persists a mapping and caches it. The cache is updated even
// when the write fails, otherwise a broken DB would make the bot spawn a new
// thread on every single DM.
func rememberMapping(ctx context.Context, userID snowflake.ID, threadID snowflake.ID) {
	cacheMu.Lock()
	threadToUser[threadID] = userID
	userToThread[userID] = threadID
	cacheMu.Unlock()

	if err := upsertMapping(ctx, userID, threadID); err != nil {
		logger.Error("MP threads: persisting mapping", "user", userID.String(), "thread", threadID.String(), "error", err)
	}
}

// upsertMapping updates first and inserts only when no row matched: the
// sql/upsert ent feature is off project-wide and user_id carries a UNIQUE
// index, so a bare Create would fail forever once a user's thread is recreated.
func upsertMapping(ctx context.Context, userID snowflake.ID, threadID snowflake.ID) error {
	client := database.Default.Ent().MPThread
	updated, err := client.Update().
		Where(mpthread.UserID(userID.String())).
		SetThreadID(threadID.String()).
		Save(ctx)
	if err != nil {
		return err
	}
	if updated > 0 {
		return nil
	}
	return client.Create().
		SetUserID(userID.String()).
		SetThreadID(threadID.String()).
		Exec(ctx)
}

// forgetThread drops a mapping from both the cache and the DB. It is called
// when the thread no longer exists on Discord so the next DM recreates one.
func forgetThread(ctx context.Context, threadID snowflake.ID) {
	cacheMu.Lock()
	userID, known := threadToUser[threadID]
	if known {
		delete(threadToUser, threadID)
		if current, ok := userToThread[userID]; ok && current == threadID {
			delete(userToThread, userID)
		}
	}
	cacheMu.Unlock()

	// user_id is UNIQUE too: clearing the user's row as well guarantees the
	// mapping for the replacement thread cannot collide with a leftover.
	stale := mpthread.ThreadID(threadID.String())
	if known {
		stale = mpthread.Or(stale, mpthread.UserID(userID.String()))
	}
	if _, err := database.Default.Ent().MPThread.Delete().Where(stale).Exec(ctx); err != nil {
		logger.Error("MP threads: deleting stale mapping", "thread", threadID.String(), "error", err)
	}
}

// lockUser returns the unlock func of the conversation lock. It guards both the
// find-or-create path and the relay order; see dispatch for the contract.
func lockUser(userID snowflake.ID) func() {
	mu := &userLocks[uint64(userID)%userLockShards]
	mu.Lock()
	return mu.Unlock
}
