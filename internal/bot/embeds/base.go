package embeds

import (
	"time"

	"github.com/disgoorg/disgo/discord"
)

var avatarURL string

func Init(url string) {
	avatarURL = url
}

func BaseEmbed() discord.Embed {
	t := time.Now()
	return discord.Embed{
		Color:     0xFFFFFF,
		Timestamp: &t,
		Footer: &discord.EmbedFooter{
			Text:    "Eve – Toujours prête à vous aider.",
			IconURL: avatarURL,
		},
	}
}
