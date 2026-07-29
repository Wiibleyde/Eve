package mpthreads

import (
	"context"
	"sync"

	"Eve/internal/database"
	"Eve/internal/database/ent/mpthread"
	"Eve/internal/logger"

	"github.com/disgoorg/snowflake/v2"
)

var (
	cacheMu      sync.RWMutex
	threadToUser = make(map[snowflake.ID]snowflake.ID)
	userToThread = make(map[snowflake.ID]snowflake.ID)
)

const userLockShards = 64

var userLocks [userLockShards]sync.Mutex

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

func rememberMapping(ctx context.Context, userID snowflake.ID, threadID snowflake.ID) {
	cacheMu.Lock()
	threadToUser[threadID] = userID
	userToThread[userID] = threadID
	cacheMu.Unlock()

	if err := upsertMapping(ctx, userID, threadID); err != nil {
		logger.Error("MP threads: persisting mapping", "user", userID.String(), "thread", threadID.String(), "error", err)
	}
}

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

	stale := mpthread.ThreadID(threadID.String())
	if known {
		stale = mpthread.Or(stale, mpthread.UserID(userID.String()))
	}
	if _, err := database.Default.Ent().MPThread.Delete().Where(stale).Exec(ctx); err != nil {
		logger.Error("MP threads: deleting stale mapping", "thread", threadID.String(), "error", err)
	}
}

func lockUser(userID snowflake.ID) func() {
	mu := &userLocks[uint64(userID)%userLockShards]
	mu.Lock()
	return mu.Unlock
}
