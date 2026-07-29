// Package streamer tracks Twitch channels and keeps one live notification
// message per (guild, channel) up to date.
//
// # Lifecycle
//
// A poller ticks every minute and diffs the live set returned by Helix against
// its in-memory view:
//
//   - started -> post the notification (optionally pinging a role) and store
//     the message ID on the row;
//   - updated -> edit the embed when the title or the game changed; a viewer
//     count that moved on its own is only worth an edit every 5 minutes;
//   - ended   -> rewrite the embed as «Stream terminé» and clear the message ID.
//
// The stored message ID is the state flag, not the in-memory map: on restart,
// a channel that is already live with a message ID resumes editing that message
// instead of posting a duplicate notification.
//
// # Configuration
//
// TWITCH_CLIENT_ID and TWITCH_CLIENT_SECRET are required. Without them the
// feature disables itself entirely: no command is registered on Discord and the
// poller never starts.
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

// helixClient returns the shared Helix client, or nil when the credentials are
// missing. The environment is read lazily — godotenv only populates it once
// main() is running, well after package initialisation.
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

// Enabled reports whether the Twitch credentials are configured.
func Enabled() bool { return helixClient() != nil }

// warnDisabled logs the "feature off" warning exactly once, however many entry
// points notice it.
func warnDisabled() {
	warnOnce.Do(func() {
		logger.Warn("Streamer feature disabled: missing Twitch credentials",
			"env", EnvClientID+", "+EnvClientSecret,
		)
	})
}

// entClient returns the ent client, or nil when the database is not open yet.
func entClient() *ent.Client {
	if database.Default == nil {
		return nil
	}
	return database.Default.Ent()
}
