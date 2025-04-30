package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken    string
	DiscordClientId string
	BlagueApiToken  string
	OwnerId         string
	WebhookUrl      string
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
	// Charge le fichier .env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	config = &Config{
		DiscordToken:    getRequiredEnv("DISCORD_TOKEN"),
		DiscordClientId: getRequiredEnv("DISCORD_CLIENT_ID"),
		BlagueApiToken:  getRequiredEnv("BLAGUE_API_TOKEN"),
		OwnerId:         getRequiredEnv("OWNER_ID"),
		WebhookUrl:      getRequiredEnv("LOGS_WEBHOOK_URL"),
	}
}

func GetConfig() *Config {
	return config
}
