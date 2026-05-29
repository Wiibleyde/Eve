package embeds

import "github.com/disgoorg/disgo/discord"

func ErrorEmbed(message string) discord.Embed {
	embed := BaseEmbed()
	embed.Color = 0xFF4444
	embed.Title = "Erreur"
	embed.Description = message
	return embed
}
