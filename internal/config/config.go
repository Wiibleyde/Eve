package config

import "os"

type Config struct {
	DiscordToken string
	DatabaseURL  string
	APIPort      string
	BotOwnerID   string
}

func Load() *Config {
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "3000"
	}

	return &Config{
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		APIPort:      apiPort,
		BotOwnerID:   os.Getenv("BOT_OWNER_ID"),
	}
}
