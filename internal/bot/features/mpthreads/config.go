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

func config() (settings, bool) {
	settingsOnce.Do(loadSettings)
	return current, current.ok
}

func Enabled() bool {
	_, ok := config()
	return ok
}
