package logger

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	DebugLogger   *log.Logger
)

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
	for i := 0; i < len(content); i++ {
		if content[i] == ':' && i+1 < len(content) && content[i+1] == ' ' {
			return w.out.Write([]byte(w.prefix + content[i+2:]))
		}
	}
	return w.out.Write([]byte(w.prefix + content))
}
