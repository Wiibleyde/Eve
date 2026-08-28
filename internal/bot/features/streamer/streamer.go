package streamer

import (
	"sync"

	"Eve/internal/config"
	"Eve/internal/database"
	"Eve/internal/database/ent"
	"Eve/internal/logger"
	"Eve/internal/twitch"
)

const CommandName = "streamer"

var (
	clientOnce sync.Once
	helix      *twitch.Client

	warnOnce sync.Once
)

func helixClient() *twitch.Client {
	clientOnce.Do(func() {
		cfg := config.Get()
		if cfg.TwitchClientID == "" || cfg.TwitchClientSecret == "" {
			return
		}
		helix = twitch.New(cfg.TwitchClientID, cfg.TwitchClientSecret)
	})
	return helix
}

func Enabled() bool { return helixClient() != nil }

func warnDisabled() {
	warnOnce.Do(func() {
		logger.Warn("Streamer feature disabled: missing Twitch credentials",
			"env", config.EnvTwitchClientID+", "+config.EnvTwitchClientSecret,
		)
	})
}

func entClient() *ent.Client {
	if database.Default == nil {
		return nil
	}
	return database.Default.Ent()
}
