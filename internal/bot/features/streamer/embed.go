package streamer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"Eve/internal/bot/embeds"
	"Eve/internal/twitch"

	"github.com/disgoorg/disgo/discord"
)

const (
	twitchColor = 0x9146FF
	endedColor  = 0x5C5C5C

	thumbnailWidth  = 1280
	thumbnailHeight = 720

	embedTitleLimit = 256
)

func channelURL(login string) string {
	return "https://twitch.tv/" + login
}

func liveEmbed(s twitch.Stream, user twitch.User, hasUser bool) discord.Embed {
	name := s.Name()
	if hasUser && user.Name() != "" {
		name = user.Name()
	}
	login := s.UserLogin
	if login == "" && hasUser {
		login = user.Login
	}

	embed := embeds.BaseEmbed()
	embed.Color = twitchColor
	embed.Title = truncate(orDefault(s.Title, "Stream en cours"), embedTitleLimit)
	embed.URL = channelURL(login)
	embed.Description = fmt.Sprintf("**%s** est en live sur Twitch !", name)

	author := &discord.EmbedAuthor{Name: name, URL: channelURL(login)}
	if hasUser {
		author.IconURL = user.ProfileImageURL
	}
	embed.Author = author

	inline := true
	embed.Fields = []discord.EmbedField{
		{Name: "Jeu", Value: orDefault(s.GameName, "Inconnu"), Inline: &inline},
		{Name: "Spectateurs", Value: strconv.Itoa(s.ViewerCount), Inline: &inline},
		{Name: "En live depuis", Value: relativeTimestamp(s.StartedAt), Inline: &inline},
	}

	if url := cacheBust(s.Thumbnail(thumbnailWidth, thumbnailHeight)); url != "" {
		embed.Image = &discord.EmbedResource{URL: url}
	}
	return embed
}

func endedEmbed(st *trackState, login string, user twitch.User, hasUser bool) discord.Embed {
	name := login
	if hasUser && user.Name() != "" {
		name = user.Name()
	}

	embed := embeds.BaseEmbed()
	embed.Color = endedColor
	embed.Title = "Stream terminé"
	embed.URL = channelURL(login)
	embed.Description = fmt.Sprintf("**%s** n'est plus en live.", name)

	author := &discord.EmbedAuthor{Name: name, URL: channelURL(login)}
	if hasUser {
		author.IconURL = user.ProfileImageURL
	}
	embed.Author = author

	inline := true
	fields := make([]discord.EmbedField, 0, 3)
	if st != nil && st.title != "" {
		fields = append(fields, discord.EmbedField{
			Name:  "Dernier titre",
			Value: truncate(st.title, 1024),
		})
	}
	if st != nil && st.game != "" {
		fields = append(fields, discord.EmbedField{Name: "Jeu", Value: st.game, Inline: &inline})
	}
	if st != nil && !st.startedAt.IsZero() {
		fields = append(fields, discord.EmbedField{
			Name:   "Durée",
			Value:  formatDuration(time.Since(st.startedAt)),
			Inline: &inline,
		})
	}
	embed.Fields = fields

	if hasUser && user.OfflineImageURL != "" {
		embed.Image = &discord.EmbedResource{URL: user.OfflineImageURL}
	}
	return embed
}

func cacheBust(url string) string {
	if url == "" {
		return ""
	}
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	return url + sep + "t=" + strconv.FormatInt(time.Now().Unix(), 10)
}

func relativeTimestamp(t time.Time) string {
	if t.IsZero() {
		return "à l'instant"
	}
	return fmt.Sprintf("<t:%d:R>", t.Unix())
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Second)

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	switch {
	case hours > 0:
		return fmt.Sprintf("%dh %02dmin", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%dmin", minutes)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

func orDefault(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

func truncate(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit-1]) + "…"
}
