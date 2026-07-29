package streamer

import (
	"os"
	"strings"
	"sync"

	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/logger"
	"Eve/internal/twitch"
)

const (
	EnvClientID     = "TWITCH_CLIENT_ID"
	EnvClientSecret = "TWITCH_CLIENT_SECRET"
)

const CommandName = "streamer"

var (
	clientOnce sync.Once
	helix      *twitch.Client

	warnOnce sync.Once
)

func helixClient() *twitch.Client {
	clientOnce.Do(func() {
		id := strings.TrimSpace(os.Getenv(EnvClientID))
		secret := strings.TrimSpace(os.Getenv(EnvClientSecret))
		if id == "" || secret == "" {
			return
		}
		helix = twitch.New(id, secret)
	})
	return helix
}

func Enabled() bool { return helixClient() != nil }

func warnDisabled() {
	warnOnce.Do(func() {
		logger.Warn("Streamer feature disabled: missing Twitch credentials",
			"env", EnvClientID+", "+EnvClientSecret,
		)
	})
}

func entClient() *ent.Client {
	if database.Default == nil {
		return nil
	}
	return database.Default.Ent()
}
