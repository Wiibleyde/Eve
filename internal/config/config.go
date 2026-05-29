package config

import "os"

type Config struct {
	DiscordToken string
	DatabaseURL  string
}

func Load() *Config {
	guildIDsStr := os.Getenv("GUILD_IDS")
	var guildIDs []string
	for _, id := range splitCSV(guildIDsStr) {
		if id != "" {
			guildIDs = append(guildIDs, id)
		}
	}
	return &Config{
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
	}
}

func splitCSV(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}
