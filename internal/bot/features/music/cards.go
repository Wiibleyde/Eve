package music

import (
	"fmt"
	"strings"

	"Eve/internal/audio"
	"Eve/internal/bot/ui"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgolink/v4/lavalink"
)

const (
	accentMusic    = 0x1DB954
	progressLength = 20
	queuePreview   = 5
	maxTitleLength = 90
)

func trackTitle(media audio.Media) string {
	title := media.Title
	if len([]rune(title)) > maxTitleLength {
		title = string([]rune(title)[:maxTitleLength-1]) + "…"
	}
	if media.URI == "" {
		return title
	}
	return fmt.Sprintf("[%s](%s)", title, media.URI)
}

func trackLine(media audio.Media) string {
	line := trackTitle(media)
	if media.Author != "" {
		line += " — " + media.Author
	}
	return line
}

func formatPosition(d lavalink.Duration) string {
	if d < 0 {
		d = 0
	}
	if hours := d.Hours(); hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, d.MinutesPart(), d.SecondsPart())
	}
	return fmt.Sprintf("%d:%02d", d.Minutes(), d.SecondsPart())
}

func progressBar(position lavalink.Duration, length lavalink.Duration) string {
	if length <= 0 {
		return strings.Repeat("▬", progressLength)
	}

	filled := int(int64(position) * progressLength / int64(length))
	if filled < 0 {
		filled = 0
	}
	if filled >= progressLength {
		filled = progressLength - 1
	}
	return strings.Repeat("▬", filled) + "🔘" + strings.Repeat("▬", progressLength-filled-1)
}

func controlRow(autoplay bool) []discord.InteractiveComponent {
	sparkle := discord.NewSecondaryButton("", CustomIDAutoplay)
	if autoplay {
		sparkle = discord.NewSuccessButton("", CustomIDAutoplay)
	}

	return []discord.InteractiveComponent{
		discord.NewPrimaryButton("", CustomIDBack).WithEmoji(discord.NewComponentEmoji("⏪")),
		discord.NewPrimaryButton("", CustomIDSkip).WithEmoji(discord.NewComponentEmoji("⏩")),
		discord.NewDangerButton("", CustomIDPlayPause).WithEmoji(discord.NewComponentEmoji("⏯️")),
		discord.NewDangerButton("", CustomIDLoop).WithEmoji(discord.NewComponentEmoji("🔁")),
		sparkle.WithEmoji(discord.NewComponentEmoji("✨")),
	}
}

func nowPlayingCard(media audio.Media, position lavalink.Duration, repeat RepeatMode, paused bool, queued int, autoplay bool) *ui.Card {
	header := "🎵 Lecture en cours"
	if paused {
		header = "⏸️ En pause"
	}

	card := ui.New().Accent(accentMusic).Title(header).Text(trackLine(media))

	if media.ArtworkURL != "" {
		card.Thumbnail(media.ArtworkURL)
	}

	if media.IsStream {
		card.Subtext("🔴 Direct")
	} else {
		card.Subtextf("`%s` %s `%s`", formatPosition(position), progressBar(position, media.Length), formatPosition(media.Length))
	}

	details := fmt.Sprintf("Boucle : **%s**", repeat.Label())
	if queued > 0 {
		details += fmt.Sprintf(" • **%d** en file d'attente", queued)
	}
	if autoplay {
		details += " • ✨ Mode automatique"
	}
	card.Divider().Subtext(details)

	return card.Row(controlRow(autoplay)...)
}

func queueCard(tracks []audio.Media, current *audio.Media, repeat RepeatMode) *ui.Card {
	card := ui.New().Accent(accentMusic).Title("📜 File d'attente")

	if current != nil {
		card.Textf("**En cours :** %s", trackLine(*current)).Divider()
	}

	shown := tracks
	if len(shown) > queuePreview {
		shown = shown[:queuePreview]
	}

	lines := make([]string, 0, len(shown))
	for i, track := range shown {
		lines = append(lines, fmt.Sprintf("**%d.** %s", i+1, trackLine(track)))
	}
	card.Text(strings.Join(lines, "\n"))

	summary := fmt.Sprintf("%d musique(s) en attente", len(tracks))
	if rest := len(tracks) - len(shown); rest > 0 {
		summary += fmt.Sprintf(" • %d non affichée(s)", rest)
	}
	if icon := repeat.Icon(); icon != "" {
		summary = icon + " " + summary
	}

	return card.Divider().Subtext(summary)
}
