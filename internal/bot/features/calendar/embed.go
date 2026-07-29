package calendar

import (
	"fmt"
	"strings"
	"time"

	"Eve/internal/bot/embeds"

	"github.com/disgoorg/disgo/discord"
)

const (
	embedTitle = "📅 Calendrier"
	// maxOngoing / maxUpcoming keep the embed readable and inside the 1024
	// character budget of a single embed field.
	maxOngoing  = 5
	maxUpcoming = 5
	// maxFieldLength is the Discord limit for an embed field value.
	maxFieldLength = 1024
	// maxSummaryLength keeps one line from swallowing the whole field.
	maxSummaryLength  = 100
	maxLocationLength = 60

	fieldOngoing  = "En cours"
	fieldUpcoming = "À venir"

	msgNoEvents      = "Aucun événement en cours ou à venir."
	footerLastUpdate = "Dernière actualisation"
)

// buildEmbed renders the persistent calendar message.
//
// now is passed in rather than read from the clock so the embed and the
// scheduled-event pass of a single tick agree on what "now" means.
func buildEmbed(events []Event, now time.Time) discord.Embed {
	ongoing, upcoming := splitEvents(events, now)

	embed := embeds.BaseEmbed()
	embed.Title = embedTitle
	// The footer icon comes from BaseEmbed; only the text changes so the
	// message reads as "last refreshed <timestamp>".
	if embed.Footer != nil {
		embed.Footer.Text = footerLastUpdate
	}
	refreshedAt := now
	embed.Timestamp = &refreshedAt

	if len(ongoing) == 0 && len(upcoming) == 0 {
		embed.Description = msgNoEvents
		return embed
	}

	if len(ongoing) > 0 {
		embed.Fields = append(embed.Fields, discord.EmbedField{
			Name:  fieldOngoing,
			Value: renderLines(ongoing),
		})
	}
	if len(upcoming) > 0 {
		embed.Fields = append(embed.Fields, discord.EmbedField{
			Name:  fieldUpcoming,
			Value: renderLines(upcoming),
		})
	}
	return embed
}

// splitEvents partitions events into the ones running right now (start ≤ now ≤
// end, both bounds inclusive per the spec) and the next ones to start. Input is
// expected sorted by start time (fetchEvents does it).
func splitEvents(events []Event, now time.Time) (ongoing []Event, upcoming []Event) {
	for _, event := range events {
		switch {
		case !event.Start.After(now) && !event.End.Before(now):
			if len(ongoing) < maxOngoing {
				ongoing = append(ongoing, event)
			}
		case event.Start.After(now):
			if len(upcoming) < maxUpcoming {
				upcoming = append(upcoming, event)
			}
		}
		if len(ongoing) >= maxOngoing && len(upcoming) >= maxUpcoming {
			break
		}
	}
	return ongoing, upcoming
}

// renderLines formats event lines, stopping before the field length limit
// instead of letting Discord reject the whole message.
func renderLines(events []Event) string {
	var builder strings.Builder
	for _, event := range events {
		line := renderLine(event)
		if builder.Len()+len(line)+1 > maxFieldLength {
			break
		}
		if builder.Len() > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(line)
	}
	if builder.Len() == 0 {
		return msgNoEvents
	}
	return builder.String()
}

func renderLine(event Event) string {
	line := fmt.Sprintf("**%s** — <t:%d:R>", inline(event.Summary, maxSummaryLength), event.Start.Unix())
	if event.AllDay {
		line += " *(journée entière)*"
	}
	if location := inline(event.Location, maxLocationLength); location != "" {
		line += " · 📍 " + location
	}
	return line
}

// inline flattens a possibly multi-line ICS value into one embed line.
func inline(value string, limit int) string {
	flattened := strings.Join(strings.Fields(strings.ReplaceAll(value, "\n", " ")), " ")
	return truncate(flattened, limit)
}

// truncate cuts on rune boundaries and marks the cut with an ellipsis.
func truncate(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	if limit == 1 {
		return "…"
	}
	return strings.TrimSpace(string(runes[:limit-1])) + "…"
}
