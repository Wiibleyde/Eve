package embeds

import "github.com/disgoorg/disgo/discord"

func SuccessEmbed(message string) discord.Embed {
	embed := BaseEmbed()
	embed.Color = 0x00FF00
	embed.Title = "Succès"
	embed.Description = message
	return embed
}
