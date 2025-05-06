package logger

import (
	"io"
	"log"
	"main/pkg/config"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	DebugLogger   *log.Logger
)

var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(input string) string {
	return ansiEscapeRegex.ReplaceAllString(input, "")
}

func getDate() string {
	dt := time.Now()
	return dt.Format("2006-01-02")
}

// EmbedColor defines Discord embed colors
type EmbedColor int

const (
	ColorWarning EmbedColor = 0xFC9803 // Yellow
	ColorInfo    EmbedColor = 0x03B1FC // Blue
	ColorError   EmbedColor = 0xFF0000 // Red
	ColorDebug   EmbedColor = 0x00FF00 // Green
	ColorDefault EmbedColor = 0x7289DA // Discord Blurple
)

func InitLogger() {
	if _, err := os.Stat("logs/"); os.IsNotExist(err) {
		os.Mkdir("logs/", 0777)
	}
	file, err := os.OpenFile("logs/logs-"+getDate()+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	fileOutput := io.Writer(file)

	fileWarningPrefix := "WARNING: "
	fileInfoPrefix := "INFO: "
	fileErrorPrefix := "ERROR: "
	fileDebugPrefix := "DEBUG: "

	consoleWarningPrefix := color.YellowString("WARNING: ")
	consoleInfoPrefix := color.BlueString("INFO: ")
	consoleErrorPrefix := color.RedString("ERROR: ")
	consoleDebugPrefix := color.GreenString("DEBUG: ")

	WarningLogger = log.New(io.MultiWriter(
		fileOutput,
		&prefixWriter{out: color.Output, prefix: consoleWarningPrefix, level: "WARNING", embedColor: ColorWarning}),
		fileWarningPrefix,
		log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger = log.New(io.MultiWriter(
		fileOutput,
		&prefixWriter{out: color.Output, prefix: consoleInfoPrefix, level: "INFO", embedColor: ColorInfo}),
		fileInfoPrefix,
		log.Ldate|log.Ltime|log.Lshortfile)

	ErrorLogger = log.New(io.MultiWriter(
		fileOutput,
		&prefixWriter{out: color.Output, prefix: consoleErrorPrefix, level: "ERROR", embedColor: ColorError}),
		fileErrorPrefix,
		log.Ldate|log.Ltime|log.Lshortfile)

	DebugLogger = log.New(io.MultiWriter(
		fileOutput,
		&prefixWriter{out: color.Output, prefix: consoleDebugPrefix, level: "DEBUG", embedColor: ColorDebug}),
		fileDebugPrefix,
		log.Ldate|log.Ltime|log.Lshortfile)
}

type prefixWriter struct {
	out        io.Writer
	prefix     string
	level      string
	embedColor EmbedColor
}

// Write writes the log message to the output stream with the specified prefix.
func (w *prefixWriter) Write(p []byte) (n int, err error) {
	content := string(p)
	finalContent := content

	// Process the content to remove redundant prefixes
	for i := 0; i < len(content); i++ {
		if content[i] == ':' && i+1 < len(content) && content[i+1] == ' ' {
			finalContent = content[i+1:]
			break
		}
	}

	// Log to Discord if configured
	if config.GetConfig().WebhookUrl != "" {
		LogToDiscord(finalContent, w.level, w.embedColor)
	}

	return w.out.Write([]byte(w.prefix + finalContent))
}

// formatLogMessage formats the message for Discord embeds
func formatLogMessage(message string) string {
	// Ensure the message isn't too long for Discord
	if len(message) > 2000 {
		message = message[:1997] + "..."
	}

	// Strip any control characters that would break Discord formatting
	message = strings.ReplaceAll(message, "\r", "")

	// Remove ANSI color codes
	message = stripANSI(message)

	return message
}

// LogToDiscord sends a message to Discord webhook with proper formatting
func LogToDiscord(message, level string, embedColor EmbedColor) {
	webhookURL := config.GetConfig().WebhookUrl
	if webhookURL == "" {
		return
	}

	// Extract webhook ID and token from URL
	parts := strings.Split(webhookURL, "/")
	if len(parts) < 7 {
		log.Printf("Invalid webhook URL format")
		return
	}

	webhookID := parts[len(parts)-2]
	webhookToken := parts[len(parts)-1]

	// Initialize Discord session (no token needed for webhooks)
	dg, err := discordgo.New("")
	if err != nil {
		log.Printf("Error creating Discord session: %s", err)
		return
	}

	formattedMsg := formatLogMessage(message)

	// Create a rich embed with more options
	embed := &discordgo.MessageEmbed{
		Title:       level,
		Description: formattedMsg,
		Color:       int(embedColor),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Eve System Log",
			IconURL: "https://static.wikia.nocookie.net/disneyemojiblitz/images/e/ed/EmojiBlitzEVE1.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Add an author field
	embed.Author = &discordgo.MessageEmbedAuthor{
		Name:    "Eve Logger",
		IconURL: "https://static.wikia.nocookie.net/disneyemojiblitz/images/e/ed/EmojiBlitzEVE1.png",
	}

	// Prepare webhook parameters
	webhookParams := &discordgo.WebhookParams{
		Username: "Eve Logger",
		Embeds:   []*discordgo.MessageEmbed{embed},
	}

	// Execute the webhook
	_, err = dg.WebhookExecute(webhookID, webhookToken, false, webhookParams)
	if err != nil {
		log.Printf("Error sending message to discord: %s", err)
	}
}
