package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken string
}

var config *Config

func InitConfig() {
	// Charge le fichier .env
	err := godotenv.Load()
	if err != nil {
		// Crée un fichier .env par défaut si inexistant
		CreateFile()
		err = godotenv.Load()
		if err != nil {
			fmt.Println("Error loading .env file:", err)
			return
		}
	}

	config = &Config{}

	// Récupère la variable depuis l'environnement
	config.BotToken = os.Getenv("BOT_TOKEN")
}

func CreateFile() {
	envContent := "BOT_TOKEN=\n"

	err := os.WriteFile(".env", []byte(envContent), 0644)
	if err != nil {
		fmt.Println("Error creating .env file:", err)
	}
}

func GetConfig() *Config {
	return config
}
