package mpthreads

import (
	"sync"

	"Eve/internal/config"
	"Eve/internal/logger"

	"github.com/disgoorg/snowflake/v2"
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
	cfg := config.Get()
	rawGuild := cfg.HomeGuildID
	rawChannel := cfg.MPChannelID

	if rawGuild == "" && rawChannel == "" {
		logger.Info("MP threads disabled: " + config.EnvHomeGuild + " and " + config.EnvMPChannel + " are unset")
		return
	}
	if rawGuild == "" || rawChannel == "" {
		logger.Warn("MP threads disabled: both environment variables are required",
			config.EnvHomeGuild, rawGuild, config.EnvMPChannel, rawChannel)
		return
	}

	guildID, err := snowflake.Parse(rawGuild)
	if err != nil {
		logger.Warn("MP threads disabled: "+config.EnvHomeGuild+" is not a valid snowflake", "value", rawGuild, "error", err)
		return
	}
	channelID, err := snowflake.Parse(rawChannel)
	if err != nil {
		logger.Warn("MP threads disabled: "+config.EnvMPChannel+" is not a valid snowflake", "value", rawChannel, "error", err)
		return
	}

	current = settings{guildID: guildID, channelID: channelID, ok: true}
}

func bridgeSettings() (settings, bool) {
	settingsOnce.Do(loadSettings)
	return current, current.ok
}

func Enabled() bool {
	_, ok := bridgeSettings()
	return ok
}
