package helpers

import (
	"os"
	"strings"
	"sync"

	"Eve/internal/logger"

	"github.com/disgoorg/snowflake/v2"
)

var (
	ownerMu     sync.Mutex
	ownerLoaded bool
	ownerID     snowflake.ID
	ownerValid  bool
)

// SetOwner installs the owner snowflake coming from the configuration
// (config.BotOwnerID). Call it during startup, before the first interaction is
// dispatched; without it the owner is resolved lazily from BOT_OWNER_ID.
func SetOwner(raw string) {
	ownerMu.Lock()
	defer ownerMu.Unlock()
	loadOwnerLocked(raw)
}

// owner resolves the owner on first use, so that godotenv has already populated
// the environment by the time BOT_OWNER_ID is read.
func owner() (snowflake.ID, bool) {
	ownerMu.Lock()
	defer ownerMu.Unlock()
	if !ownerLoaded {
		loadOwnerLocked(os.Getenv("BOT_OWNER_ID"))
	}
	return ownerID, ownerValid
}

func loadOwnerLocked(raw string) {
	ownerLoaded = true
	ownerID, ownerValid = 0, false

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	id, err := snowflake.Parse(raw)
	if err != nil {
		logger.Warn("BOT_OWNER_ID is not a valid snowflake, owner checks disabled", "value", raw, "error", err)
		return
	}
	ownerID, ownerValid = id, true
}

// IsOwner reports whether the given user is the bot owner. It returns false when
// no owner is configured, so owner-only paths stay closed by default.
func IsOwner(userID snowflake.ID) bool {
	id, ok := owner()
	return ok && userID == id
}

// OwnerConfigured reports whether a valid owner is configured. Owner-only
// features use it to disable themselves instead of being unusable.
func OwnerConfigured() bool {
	_, ok := owner()
	return ok
}

func OwnerID() (snowflake.ID, bool) {
	return owner()
}
