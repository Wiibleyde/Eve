package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorGray   = "\033[90m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// LevelEvent is a custom log level for incoming Discord interactions (commands, buttons).
const LevelEvent = slog.Level(slog.LevelInfo + 1)

type prettyHandler struct {
	mu  sync.Mutex
	out io.Writer
	lvl slog.Level
}

func (h *prettyHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.lvl
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	var buf bytes.Buffer

	buf.WriteString(colorGray)
	buf.WriteString(r.Time.Format(time.TimeOnly))
	buf.WriteString(colorReset)
	buf.WriteByte(' ')

	switch r.Level {
	case slog.LevelDebug:
		fmt.Fprintf(&buf, "%sDEBUG%s", colorBlue, colorReset)
	case slog.LevelInfo:
		fmt.Fprintf(&buf, "%s INFO%s", colorGreen, colorReset)
	case LevelEvent:
		fmt.Fprintf(&buf, "%sEVENT%s", colorCyan, colorReset)
	case slog.LevelWarn:
		fmt.Fprintf(&buf, "%s WARN%s", colorYellow, colorReset)
	case slog.LevelError:
		fmt.Fprintf(&buf, "%sERROR%s", colorRed, colorReset)
	default:
		fmt.Fprintf(&buf, "%s%5s%s", colorBold, r.Level.String(), colorReset)
	}
	buf.WriteByte(' ')

	fmt.Fprintf(&buf, "%s%s%s", colorBold, r.Message, colorReset)

	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(&buf, "  %s%s%s=%v", colorGray, a.Key, colorReset, a.Value)
		return true
	})

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(buf.Bytes())
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *prettyHandler) WithGroup(name string) slog.Handler       { return h }

var l *slog.Logger

func init() {
	l = slog.New(&prettyHandler{
		out: os.Stdout,
		lvl: slog.LevelDebug,
	})
}

func Debug(msg string, args ...any) { l.Debug(msg, args...) }
func Info(msg string, args ...any)  { l.Info(msg, args...) }
func Event(msg string, args ...any) { l.Log(context.Background(), LevelEvent, msg, args...) }
func Warn(msg string, args ...any)  { l.Warn(msg, args...) }
func Error(msg string, args ...any) { l.Error(msg, args...) }

func Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}
