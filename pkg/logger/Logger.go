package logger

import (
	"fmt"
	"io"
	"log"
	"main/pkg/config"
	"os"
	"regexp"
	"time"

	"github.com/fatih/color"
	"github.com/gtuk/discordwebhook"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	DebugLogger   *log.Logger
)

var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// ptrToString converts a string to a pointer to that string.
func ptrToString(s string) *string {
	return &s
}

func stripANSI(input string) string {
	return ansiEscapeRegex.ReplaceAllString(input, "")
}

func getDate() string {
	dt := time.Now()
	return dt.Format("2006-01-02")
}

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
		&prefixWriter{out: color.Output, prefix: consoleWarningPrefix}),
		fileWarningPrefix,
		log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger = log.New(io.MultiWriter(
		fileOutput,
		&prefixWriter{out: color.Output, prefix: consoleInfoPrefix}),
		fileInfoPrefix,
		log.Ldate|log.Ltime|log.Lshortfile)

	ErrorLogger = log.New(io.MultiWriter(
		fileOutput,
		&prefixWriter{out: color.Output, prefix: consoleErrorPrefix}),
		fileErrorPrefix,
		log.Ldate|log.Ltime|log.Lshortfile)

	DebugLogger = log.New(io.MultiWriter(
		fileOutput,
		&prefixWriter{out: color.Output, prefix: consoleDebugPrefix}),
		fileDebugPrefix,
		log.Ldate|log.Ltime|log.Lshortfile)
}

type prefixWriter struct {
	out    io.Writer
	prefix string
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
		LogToDiscord(finalContent, w.prefix)
	}

	return w.out.Write([]byte(w.prefix + finalContent))
}

func LogToDiscord(message string, level string) {
	var username = "Panel"
	var url = config.GetConfig().WebhookUrl

	if len(message) > 1900 {
		message = message[:1900] + "..."
	}

	cleanLevel := stripANSI(level)
	if len(cleanLevel) > 0 && cleanLevel[len(cleanLevel)-1] == ':' {
		cleanLevel = cleanLevel[:len(cleanLevel)-1]
	}

	embed := discordwebhook.Embed{
		Title:       &cleanLevel,
		Description: &message,
		Color:       ptrToString(fmt.Sprintf("%d", getEmbedColor(cleanLevel))),
	}

	messageType := discordwebhook.Message{
		Username: &username,
		Embeds:   &[]discordwebhook.Embed{embed},
	}

	err := discordwebhook.SendMessage(url, messageType)
	if err != nil {
		// Don't use ErrorLogger here to avoid infinite recursion
		log.Printf("Error sending message to discord: %s", err)
	}
}

func getEmbedColor(level string) int {
	switch level {
	case "WARNING: ":
		return 16776960 // Yellow
	case "INFO: ":
		return 255 // Blue
	case "ERROR: ":
		return 16711680 // Red
	case "DEBUG: ":
		return 65280 // Green
	default:
		return 0 // Default color
	}
}
