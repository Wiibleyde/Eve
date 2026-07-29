package ping

import (
	"fmt"
	"os"
	"time"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/shirou/gopsutil/v3/process"
)

const (
	PING_THRESHOLD_VERY_GOOD = 50
	PING_THRESHOLD_GOOD      = 100
	PING_THRESHOLD_CORRECT   = 150
	PING_THRESHOLD_WEAK      = 200
	PING_THRESHOLD_BAD       = 500
)

var startTime = time.Now()

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "ping",
		Description: "Récupèrer le status du bot",
	},
}

func pingColor(ms int64) int {
	switch {
	case ms <= PING_THRESHOLD_VERY_GOOD:
		return 0x57F287
	case ms <= PING_THRESHOLD_GOOD:
		return 0x7FFF00
	case ms <= PING_THRESHOLD_CORRECT:
		return 0xFEE75C
	case ms <= PING_THRESHOLD_WEAK:
		return 0xFF8C00
	case ms <= PING_THRESHOLD_BAD:
		return 0xED4245
	default:
		return 0x8B0000
	}
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	if days > 0 {
		return fmt.Sprintf("%dj %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	latencyMs := e.Client().Gateway.Latency().Milliseconds()

	proc, _ := process.NewProcess(int32(os.Getpid()))

	var memMB float64
	if memInfo, err := proc.MemoryInfo(); err == nil {
		memMB = float64(memInfo.RSS) / 1024 / 1024
	}

	var cpuPct float64
	if pct, err := proc.CPUPercent(); err == nil {
		cpuPct = pct
	}

	card := ui.New().
		Accent(pingColor(latencyMs)).
		Title("🏓 Pong !").
		Fields(
			ui.Field{Name: "Latence", Value: fmt.Sprintf("`%d ms`", latencyMs), Inline: true},
			ui.Field{Name: "Mémoire", Value: fmt.Sprintf("`%.1f Mo`", memMB), Inline: true},
			ui.Field{Name: "CPU", Value: fmt.Sprintf("`%.1f %%`", cpuPct), Inline: true},
			ui.Field{Name: "Uptime", Value: fmt.Sprintf("`%s`", formatUptime(time.Since(startTime))), Inline: true},
		)

	helpers.RespondEphemeralCard(e, card)
}
