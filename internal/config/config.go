package config

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"Eve/internal/logger"
)

const (
	EnvDiscordToken       = "DISCORD_TOKEN"
	EnvDatabaseURL        = "DATABASE_URL"
	EnvAPIPort            = "API_PORT"
	EnvBotOwnerID         = "BOT_OWNER_ID"
	EnvHomeGuild          = "EVE_HOME_GUILD"
	EnvMPChannel          = "MP_CHANNEL"
	EnvTwitchClientID     = "TWITCH_CLIENT_ID"
	EnvTwitchClientSecret = "TWITCH_CLIENT_SECRET"
	EnvBlagueAPIToken     = "BLAGUE_API_TOKEN"
	EnvMotusAPIURL        = "MOTUS_API_URL"
	EnvMotusWordsFile     = "MOTUS_WORDS_FILE"
	EnvOllamaURL          = "OLLAMA_URL"
	EnvOllamaModel        = "OLLAMA_MODEL"
	EnvOllamaThreads      = "OLLAMA_NUM_THREADS"
	EnvSearxngURL         = "SEARXNG_URL"
	EnvLavalinkAddress    = "LAVALINK_ADDRESS"
	EnvLavalinkPassword   = "LAVALINK_PASSWORD"
	EnvLavalinkSecure     = "LAVALINK_SECURE"
	EnvYtDlpPath          = "YTDLP_PATH"
	EnvYtDlpJSRuntime     = "YTDLP_JS_RUNTIME"
)

const (
	DefaultAPIPort       = "3000"
	DefaultMotusAPIURL   = "https://trouve-mot.fr/api/random"
	DefaultOllamaModel   = "llama3.2:1b"
	DefaultOllamaThreads = 3
)

type Config struct {
	DiscordToken string
	DatabaseURL  string
	APIPort      string
	BotOwnerID   string

	HomeGuildID string
	MPChannelID string

	TwitchClientID     string
	TwitchClientSecret string

	BlagueAPIToken string

	MotusAPIURL    string
	MotusWordsFile string

	OllamaURL     string
	OllamaModel   string
	OllamaThreads int

	SearxngURL string

	LavalinkAddress  string
	LavalinkPassword string
	LavalinkSecure   bool

	YtDlpPath      string
	YtDlpJSRuntime string
}

var (
	once    sync.Once
	current *Config
)

func Get() *Config {
	once.Do(func() { current = load() })
	return current
}

func load() *Config {
	return &Config{
		DiscordToken:       value(EnvDiscordToken),
		DatabaseURL:        value(EnvDatabaseURL),
		APIPort:            valueOr(EnvAPIPort, DefaultAPIPort),
		BotOwnerID:         value(EnvBotOwnerID),
		HomeGuildID:        value(EnvHomeGuild),
		MPChannelID:        value(EnvMPChannel),
		TwitchClientID:     value(EnvTwitchClientID),
		TwitchClientSecret: value(EnvTwitchClientSecret),
		BlagueAPIToken:     value(EnvBlagueAPIToken),
		MotusAPIURL:        valueOr(EnvMotusAPIURL, DefaultMotusAPIURL),
		MotusWordsFile:     value(EnvMotusWordsFile),
		OllamaURL:          baseURL(EnvOllamaURL),
		OllamaModel:        valueOr(EnvOllamaModel, DefaultOllamaModel),
		OllamaThreads:      positiveInt(EnvOllamaThreads, DefaultOllamaThreads),
		SearxngURL:         baseURL(EnvSearxngURL),
		LavalinkAddress:    value(EnvLavalinkAddress),
		LavalinkPassword:   value(EnvLavalinkPassword),
		LavalinkSecure:     boolean(EnvLavalinkSecure),
		YtDlpPath:          value(EnvYtDlpPath),
		YtDlpJSRuntime:     value(EnvYtDlpJSRuntime),
	}
}

func value(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func valueOr(key string, fallback string) string {
	if v := value(key); v != "" {
		return v
	}
	return fallback
}

func boolean(key string) bool {
	switch strings.ToLower(value(key)) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

func baseURL(key string) string {
	return strings.TrimRight(value(key), "/")
}

func positiveInt(key string, fallback int) int {
	raw := value(key)
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < 1 {
		logger.Warn("Ignoring invalid "+key, "value", raw)
		return fallback
	}
	return parsed
}
