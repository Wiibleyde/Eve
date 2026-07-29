package calendar

import (
	"fmt"
	"strings"
	"time"

	"Eve/internal/bot/ui"
)

const (
	cardTitle         = "📅 Calendrier"
	calendarColor     = 0x3BA55D
	maxOngoing        = 5
	maxUpcoming       = 5
	maxSectionLength  = 900
	maxSummaryLength  = 100
	maxLocationLength = 60

	sectionOngoing  = "🟢 En cours"
	sectionUpcoming = "🔜 À venir"

	msgNoEvents      = "Aucun événement en cours ou à venir."
	footerLastUpdate = "Dernière actualisation"
)

func buildCard(events []Event, now time.Time) *ui.Card {
	ongoing, upcoming := splitEvents(events, now)

	card := ui.New().
		Accent(calendarColor).
		Title(cardTitle).
		Footer(footerLastUpdate)

	if len(ongoing) == 0 && len(upcoming) == 0 {
		return card.Text(msgNoEvents)
	}

	if len(ongoing) > 0 {
		card.Heading(sectionOngoing).Text(renderLines(ongoing))
	}
	if len(upcoming) > 0 {
		if len(ongoing) > 0 {
			card.Divider()
		}
		card.Heading(sectionUpcoming).Text(renderLines(upcoming))
	}
	return card
}

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

func renderLines(events []Event) string {
	var builder strings.Builder
	for _, event := range events {
		line := renderLine(event)
		if builder.Len()+len(line)+1 > maxSectionLength {
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

func inline(value string, limit int) string {
	flattened := strings.Join(strings.Fields(strings.ReplaceAll(value, "\n", " ")), " ")
	return truncate(flattened, limit)
}

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
