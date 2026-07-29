// Package userinfo exposes the two user context menus that surface a member's
// images: the profile picture and the profile banner.
//
// Both are right-click only — there is no slash command counterpart — and both
// answer ephemerally, so looking someone up never spams a channel.
package userinfo

import (
	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

// Context menu names double as their routing keys.
const (
	MenuAvatar = "Récupèrer la photo de profil"
	MenuBanner = "Récupèrer la bannière"
)

const imageSize = 1024

// User-facing strings (French, per project conventions).
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

// HandleAvatar shows the target's avatar, preferring the guild-specific one
// over the global account avatar.
func HandleAvatar(e *events.ApplicationCommandInteractionCreate) {
	data := e.UserCommandInteractionData()
	user := data.TargetUser()

	embed := embeds.BaseEmbed()
	embed.Title = titleAvatar
	embed.Description = user.Mention()
	embed.Image = &discord.EmbedResource{URL: avatarURL(e)}
	helpers.RespondEphemeralEmbed(e, embed)
}

// avatarURL resolves the avatar to display: the guild avatar when the member has
// one, the account avatar (or Discord's default) otherwise.
func avatarURL(e *events.ApplicationCommandInteractionCreate) string {
	data := e.UserCommandInteractionData()
	opts := []discord.CDNOpt{discord.WithFormat(discord.FileFormatPNG), discord.WithSize(imageSize)}

	if guildID := e.GuildID(); guildID != nil {
		member := data.TargetMember()
		// Resolved members come without a guild ID, and the member avatar route
		// needs one — fill it in from the interaction.
		member.GuildID = *guildID
		if guildAvatar := member.AvatarURL(opts...); guildAvatar != nil {
			return *guildAvatar
		}
	}
	return data.TargetUser().EffectiveAvatarURL(opts...)
}

func HandleBanner(e *events.ApplicationCommandInteractionCreate) {
	target := e.UserCommandInteractionData().TargetUser()

	// The banner is absent from interaction payloads; only a full user object
	// carries it, which means a REST round trip.
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
