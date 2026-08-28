package helpers

import (
	"strings"
	"sync"

	"Eve/internal/config"
	"Eve/internal/logger"

	"github.com/disgoorg/snowflake/v2"
)

var (
	ownerMu     sync.Mutex
	ownerLoaded bool
	ownerID     snowflake.ID
	ownerValid  bool
)

func SetOwner(raw string) {
	ownerMu.Lock()
	defer ownerMu.Unlock()
	loadOwnerLocked(raw)
}

func owner() (snowflake.ID, bool) {
	ownerMu.Lock()
	defer ownerMu.Unlock()
	if !ownerLoaded {
		loadOwnerLocked(config.Get().BotOwnerID)
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
		logger.Warn(config.EnvBotOwnerID+" is not a valid snowflake, owner checks disabled", "value", raw, "error", err)
		return
	}
	ownerID, ownerValid = id, true
}

func IsOwner(userID snowflake.ID) bool {
	id, ok := owner()
	return ok && userID == id
}

func OwnerConfigured() bool {
	_, ok := owner()
	return ok
}

func OwnerID() (snowflake.ID, bool) {
	return owner()
}
