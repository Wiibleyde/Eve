package streamer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"Eve/internal/bot/ui"
	"Eve/internal/twitch"

	"github.com/disgoorg/disgo/discord"
)

const (
	twitchColor = 0x9146FF
	endedColor  = 0x5C5C5C

	thumbnailWidth  = 1280
	thumbnailHeight = 720

	titleLimit = 200
)

func channelURL(login string) string {
	return "https://twitch.tv/" + login
}

func liveCard(s twitch.Stream, user twitch.User, hasUser bool, roleMention string) *ui.Card {
	name := s.Name()
	if hasUser && user.Name() != "" {
		name = user.Name()
	}
	login := s.UserLogin
	if login == "" && hasUser {
		login = user.Login
	}

	card := ui.New().
		Accent(twitchColor).
		Titlef("🔴 [%s](%s)", escapeLinkLabel(truncate(orDefault(s.Title, "Stream en cours"), titleLimit)), channelURL(login))

	if hasUser {
		card.Thumbnail(user.ProfileImageURL)
	}

	if roleMention != "" {
		card.Text(roleMention)
	}

	card.Textf("**%s** est en live sur Twitch !", name).
		Fields(
			ui.Field{Name: "Jeu", Value: orDefault(s.GameName, "Inconnu"), Inline: true},
			ui.Field{Name: "Spectateurs", Value: strconv.Itoa(s.ViewerCount), Inline: true},
			ui.Field{Name: "En live depuis", Value: relativeTimestamp(s.StartedAt), Inline: true},
		).
		Image(cacheBust(s.Thumbnail(thumbnailWidth, thumbnailHeight))).
		Row(discord.NewLinkButton("Regarder le stream", channelURL(login)))

	return card
}

func endedCard(st *trackState, login string, user twitch.User, hasUser bool) *ui.Card {
	name := login
	if hasUser && user.Name() != "" {
		name = user.Name()
	}

	card := ui.New().
		Accent(endedColor).
		Titlef("⚫ [Stream terminé](%s)", channelURL(login)).
		Textf("**%s** n'est plus en live.", name)

	if hasUser {
		card.Thumbnail(user.ProfileImageURL)
	}

	if st != nil && st.title != "" {
		card.Fields(ui.Field{Name: "Dernier titre", Value: truncate(st.title, titleLimit)})
	}

	stats := make([]ui.Field, 0, 2)
	if st != nil && st.game != "" {
		stats = append(stats, ui.Field{Name: "Jeu", Value: st.game, Inline: true})
	}
	if st != nil && !st.startedAt.IsZero() {
		stats = append(stats, ui.Field{Name: "Durée", Value: formatDuration(time.Since(st.startedAt)), Inline: true})
	}
	card.Fields(stats...)

	if hasUser {
		card.Image(user.OfflineImageURL)
	}

	return card.Row(discord.NewLinkButton("Voir la chaîne", channelURL(login)))
}

var linkLabelEscaper = strings.NewReplacer("[", "(", "]", ")")

func escapeLinkLabel(s string) string {
	return linkLabelEscaper.Replace(s)
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
