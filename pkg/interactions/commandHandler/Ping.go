package commandHandler

import (
	"fmt"
	"main/pkg/bot_utils"
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	PING_THRESHOLD_VERY_GOOD = 50
	PING_THRESHOLD_GOOD      = 100
	PING_THRESHOLD_CORRECT   = 150
	PING_THRESHOLD_WEAK      = 200
	PING_THRESHOLD_BAD       = 500
)

func PingHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ping := s.HeartbeatLatency().Milliseconds()
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

	embed := &discordgo.MessageEmbed{
		Title:       "Status du bot",
		Description: pingMessage,
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Valeur",
				Value:  fmt.Sprintf("%d ms", ping),
				Inline: true,
			},
			{
				Name:   "Mémoire utilisée",
				Value:  fmt.Sprintf("%.2f Mo", heapConsumed),
				Inline: true,
			},
			{
				Name:   "Uptime",
				Value:  uptimeMessage,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Eve – Toujours prête à vous aider.",
			IconURL: s.State.User.AvatarURL("64"),
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return err
	}

	return nil
}
