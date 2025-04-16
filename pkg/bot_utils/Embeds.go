package bot_utils

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func BasicEmbedBuilder(s *discordgo.Session) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Color:  0xFFFFFF,
		Fields: []*discordgo.MessageEmbedField{},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Eve – Toujours prête à vous aider.",
			IconURL: s.State.User.AvatarURL("64"),
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func SuccessEmbed(s *discordgo.Session, reason string) *discordgo.MessageEmbed {
	embed := BasicEmbedBuilder(s)
	embed.Title = "Succès !"
	embed.Description = reason
	embed.Color = 0x00FF00
	return embed
}

func ErrorEmbed(s *discordgo.Session, reason error) *discordgo.MessageEmbed {
	embed := BasicEmbedBuilder(s)
	embed.Title = "Oops ! Une erreur s'est produite"
	embed.Description = reason.Error()
	embed.Color = 0xFF0000
	return embed
}

func WarningEmbed(s *discordgo.Session, reason string) *discordgo.MessageEmbed {
	embed := BasicEmbedBuilder(s)
	embed.Title = "Attention !"
	embed.Description = reason
	embed.Color = 0xFFA500
	return embed
}

