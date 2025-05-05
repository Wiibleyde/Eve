package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken       string
	DiscordClientId    string
	BlagueApiToken     string
	OwnerId            string
	WebhookUrl         string
	GoogleApiKey       string
	TwitchClientId     string
	TwitchClientSecret string
}

var config *Config

// getRequiredEnv récupère une variable d'environnement et vérifie qu'elle n'est pas vide
func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		fmt.Printf("%s is not set\n", key)
		os.Exit(1)
	}
	return value
}

func InitConfig() {
	// Essaye de charger le fichier .env, mais continue même s'il n'existe pas
	_ = godotenv.Load() // Ignore error, will use environment variables directly if .env doesn't exist

	config = &Config{
		DiscordToken:       getRequiredEnv("DISCORD_TOKEN"),
		DiscordClientId:    getRequiredEnv("DISCORD_CLIENT_ID"),
		BlagueApiToken:     getRequiredEnv("BLAGUE_API_TOKEN"),
		OwnerId:            getRequiredEnv("OWNER_ID"),
		WebhookUrl:         getRequiredEnv("LOGS_WEBHOOK_URL"),
		GoogleApiKey:       getRequiredEnv("GOOGLE_API_KEY"),
		TwitchClientId:     getRequiredEnv("TWITCH_CLIENT_ID"),
		TwitchClientSecret: getRequiredEnv("TWITCH_CLIENT_SECRET"),
	}
}

func GetConfig() *Config {
	return config
}
