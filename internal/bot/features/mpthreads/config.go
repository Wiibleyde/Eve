// Package mpthreads bridges the bot's direct messages with staff threads.
//
// Every DM a user sends to the bot is mirrored into a dedicated public thread
// under a staff channel of the home guild; any staff message posted in that
// thread is relayed back to the user's DMs.
//
// The feature is configured entirely through the environment:
//
//	EVE_HOME_GUILD=<guild snowflake>   guild hosting the MP threads
//	MP_CHANNEL=<channel snowflake>     parent channel of the MP threads
//
// Both unset (or either invalid) disables the feature: Attach becomes a no-op.
//
// It requires the DirectMessages, GuildMessages and MessageContent intents.
package mpthreads

import (
	"os"
	"strings"
	"sync"

	"Eve/internal/logger"

	"github.com/disgoorg/snowflake/v2"
)

const (
	EnvHomeGuild = "EVE_HOME_GUILD"
	EnvMPChannel = "MP_CHANNEL"
)

type settings struct {
	guildID   snowflake.ID
	channelID snowflake.ID
	ok        bool
}

var (
	settingsOnce sync.Once
	current      settings
)

// loadSettings reads the environment exactly once, lazily so that godotenv has
// already populated it.
func loadSettings() {
	rawGuild := strings.TrimSpace(os.Getenv(EnvHomeGuild))
	rawChannel := strings.TrimSpace(os.Getenv(EnvMPChannel))

	if rawGuild == "" && rawChannel == "" {
		logger.Info("MP threads disabled: " + EnvHomeGuild + " and " + EnvMPChannel + " are unset")
		return
	}
	if rawGuild == "" || rawChannel == "" {
		logger.Warn("MP threads disabled: both environment variables are required",
			EnvHomeGuild, rawGuild, EnvMPChannel, rawChannel)
		return
	}

	guildID, err := snowflake.Parse(rawGuild)
	if err != nil {
		logger.Warn("MP threads disabled: "+EnvHomeGuild+" is not a valid snowflake", "value", rawGuild, "error", err)
		return
	}
	channelID, err := snowflake.Parse(rawChannel)
	if err != nil {
		logger.Warn("MP threads disabled: "+EnvMPChannel+" is not a valid snowflake", "value", rawChannel, "error", err)
		return
	}

	current = settings{guildID: guildID, channelID: channelID, ok: true}
}

// config returns the parsed configuration and whether the feature is usable.
func config() (settings, bool) {
	settingsOnce.Do(loadSettings)
	return current, current.ok
}

// Enabled reports whether both environment variables are set and valid.
func Enabled() bool {
	_, ok := config()
	return ok
}
