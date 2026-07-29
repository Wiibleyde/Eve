package userinfo

import (
	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const (
	MenuAvatar = "Récupèrer la photo de profil"
	MenuBanner = "Récupèrer la bannière"
)

const imageSize = 1024

const (
	titleAvatar   = "Photo de profil"
	titleBanner   = "Bannière"
	msgNoBanner   = "Cet utilisateur n'a pas de bannière."
	msgFetchError = "Impossible de récupérer cet utilisateur."
)

var Commands = []discord.ApplicationCommandCreate{
	discord.UserCommandCreate{Name: MenuAvatar},
	discord.UserCommandCreate{Name: MenuBanner},
}

func HandleAvatar(e *events.ApplicationCommandInteractionCreate) {
	data := e.UserCommandInteractionData()
	user := data.TargetUser()

	embed := embeds.BaseEmbed()
	embed.Title = titleAvatar
	embed.Description = user.Mention()
	embed.Image = &discord.EmbedResource{URL: avatarURL(e)}
	helpers.RespondEphemeralEmbed(e, embed)
}

func avatarURL(e *events.ApplicationCommandInteractionCreate) string {
	data := e.UserCommandInteractionData()
	opts := []discord.CDNOpt{discord.WithFormat(discord.FileFormatPNG), discord.WithSize(imageSize)}

	if guildID := e.GuildID(); guildID != nil {
		member := data.TargetMember()
		member.GuildID = *guildID
		if guildAvatar := member.AvatarURL(opts...); guildAvatar != nil {
			return *guildAvatar
		}
	}
	return data.TargetUser().EffectiveAvatarURL(opts...)
}

func HandleBanner(e *events.ApplicationCommandInteractionCreate) {
	target := e.UserCommandInteractionData().TargetUser()

	user, err := e.Client().Rest.GetUser(target.ID)
	if err != nil {
		logger.Error("fetching user for banner", "user", target.ID.String(), "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(msgFetchError))
		return
	}

	bannerURL := user.BannerURL(discord.WithSize(imageSize))
	if bannerURL == nil {
		helpers.RespondEphemeral(e, msgNoBanner)
		return
	}

	embed := embeds.BaseEmbed()
	embed.Title = titleBanner
	embed.Description = user.Mention()
	embed.Image = &discord.EmbedResource{URL: *bannerURL}
	helpers.RespondEphemeralEmbed(e, embed)
}
