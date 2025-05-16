package commandHandler

import (
	"fmt"
	"main/pkg/bot_utils"
	"runtime"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const (
	PING_THRESHOLD_VERY_GOOD = 50
	PING_THRESHOLD_GOOD      = 100
	PING_THRESHOLD_CORRECT   = 150
	PING_THRESHOLD_WEAK      = 200
	PING_THRESHOLD_BAD       = 500
)

func PingHandler(event *events.ApplicationCommandInteractionCreate) {
	ping := event.Client().Gateway().Latency()
	var pingMessage string
	var color int

	switch {
	case ping <= PING_THRESHOLD_VERY_GOOD:
		pingMessage = "Très bon 🟢"
		color = 0x00FF00
	case ping <= PING_THRESHOLD_GOOD:
		pingMessage = "Bon 🟢"
		color = 0x00FF00
	case ping <= PING_THRESHOLD_CORRECT:
		pingMessage = "Correct 🟡"
		color = 0xFFFF00
	case ping <= PING_THRESHOLD_WEAK:
		pingMessage = "Faible 🟠"
		color = 0xFFA500
	case ping <= PING_THRESHOLD_BAD:
		pingMessage = "Mauvais 🔴"
		color = 0xFF0000
	default:
		pingMessage = "Très mauvais 🔴"
		color = 0xFF0000
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	heapConsumed := float64(memStats.HeapAlloc) / (1024 * 1024)

	uptime := time.Since(bot_utils.StartTime)
	uptimeMessage := fmt.Sprintf("%d minutes, %d secondes", int(uptime.Minutes()), int(uptime.Seconds())%60)

	embed := discord.NewEmbedBuilder().
		SetTitle("Status du bot").
		SetDescription(pingMessage).
		SetColor(color).
		AddField("Valeur", fmt.Sprintf("%d ms", ping), true).
		AddField("Mémoire utilisée", fmt.Sprintf("%.2f Mo", heapConsumed), true).
		AddField("Uptime", uptimeMessage, true).
		SetFooter("Eve – Toujours prête à vous aider.", *event.User().AvatarURL()).
		SetTimestamp(time.Now()).
		Build()

	err := event.Respond(
		discord.InteractionResponseTypeCreateMessage,
		discord.MessageCreate{
			Embeds: []discord.Embed{embed},
		},
	)
	if err != nil {
		return
	}
}
