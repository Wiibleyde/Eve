package calendar

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"Eve/internal/bot/maintenance"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

const (
	TickInterval       = 5 * time.Minute
	startupDelay       = 30 * time.Second
	scheduledEventLead = 30 * time.Minute
	tickTimeout        = TickInterval - 30*time.Second

	maxScheduledEventName        = 100
	maxScheduledEventDescription = 1000
	maxScheduledEventLocation    = 100
	defaultEventLocation         = "Non précisé"
)

func StartScheduler(client *bot.Client) {
	go func() {
		logger.Info("Calendar scheduler started", "interval", TickInterval.String())
		time.Sleep(startupDelay)

		ticker := time.NewTicker(TickInterval)
		defer ticker.Stop()

		runTick(client)
		for range ticker.C {
			if maintenance.Enabled() {
				logger.Debug("Calendar tick skipped: maintenance mode")
				continue
			}
			runTick(client)
		}
	}()
}

func runTick(client *bot.Client) {
	if maintenance.Enabled() {
		logger.Debug("Calendar tick skipped: maintenance mode")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), tickTimeout)
	defer cancel()

	urls, err := calendarURLs(ctx)
	if err != nil {
		logger.Error("Calendar: listing configured guilds failed", "error", err)
		return
	}
	if len(urls) == 0 {
		return
	}

	logger.Debug("Calendar tick", "guilds", len(urls))
	for guildID, rawURL := range urls {
		refreshGuildSafe(ctx, client, guildID, rawURL)
	}
}

func refreshGuildSafe(ctx context.Context, client *bot.Client, guildID string, rawURL string) {
	defer func() {
		if rec := recover(); rec != nil {
			logger.Error("Calendar: panic while refreshing guild",
				"guild", guildID,
				"panic", fmt.Sprint(rec),
				"stack", strings.ReplaceAll(string(debug.Stack()), "\n", " | "),
			)
		}
	}()

	if err := refreshGuild(ctx, client, guildID, rawURL); err != nil {
		logger.Error("Calendar: guild refresh failed", "guild", guildID, "error", err)
	}
}

func refreshGuild(ctx context.Context, client *bot.Client, guildIDStr string, rawURL string) error {
	guildID, err := snowflake.Parse(guildIDStr)
	if err != nil {
		return fmt.Errorf("parsing guild id: %w", err)
	}

	parsed, err := fetchEvents(ctx, rawURL)
	if err != nil {
		return fmt.Errorf("fetching calendar: %w", err)
	}

	cfg, err := loadGuildCalendar(ctx, guildIDStr)
	if err != nil {
		return fmt.Errorf("loading calendar config: %w", err)
	}

	now := time.Now()

	if err := updateCalendarMessage(ctx, client, cfg, buildCard(parsed, now)); err != nil &&
		!errors.Is(err, errCalendarMessageGone) && !errors.Is(err, errNoCalendarMessage) {
		logger.Warn("Calendar: refreshing message failed", "guild", guildIDStr, "error", err)
	}

	return syncScheduledEvents(client, guildID, parsed, now)
}

func syncScheduledEvents(client *bot.Client, guildID snowflake.ID, parsed []Event, now time.Time) error {
	due := eventsStartingWithin(parsed, now, scheduledEventLead)
	if len(due) == 0 {
		return nil
	}

	existing, err := client.Rest.GetGuildScheduledEvents(guildID, false)
	if err != nil {
		return fmt.Errorf("listing guild scheduled events: %w", err)
	}

	for _, event := range due {
		if scheduledEventExists(existing, event) {
			continue
		}

		created, err := client.Rest.CreateGuildScheduledEvent(guildID, scheduledEventCreate(event))
		if err != nil {
			logger.Warn("Calendar: creating scheduled event failed",
				"guild", guildID.String(), "event", event.Summary, "error", err)
			continue
		}
		existing = append(existing, *created)
		logger.Info("Calendar: scheduled event created",
			"guild", guildID.String(), "event", created.Name, "start", created.ScheduledStartTime.Format(time.RFC3339))
	}
	return nil
}

func eventsStartingWithin(parsed []Event, now time.Time, lead time.Duration) []Event {
	deadline := now.Add(lead)
	due := make([]Event, 0, 4)
	for _, event := range parsed {
		if event.Start.After(now) && !event.Start.After(deadline) {
			due = append(due, event)
		}
	}
	return due
}

func scheduledEventCreate(event Event) discord.GuildScheduledEventCreate {
	end := event.End
	if !end.After(event.Start) {
		end = event.Start.Add(defaultTimedDuration)
	}

	location := inline(event.Location, maxScheduledEventLocation)
	if location == "" {
		location = defaultEventLocation
	}

	return discord.GuildScheduledEventCreate{
		Name:               scheduledEventName(event),
		Description:        truncate(event.Description, maxScheduledEventDescription),
		PrivacyLevel:       discord.ScheduledEventPrivacyLevelGuildOnly,
		ScheduledStartTime: event.Start,
		ScheduledEndTime:   &end,
		EntityType:         discord.ScheduledEventEntityTypeExternal,
		EntityMetaData:     &discord.EntityMetaData{Location: location},
	}
}

func scheduledEventName(event Event) string {
	name := inline(event.Summary, maxScheduledEventName)
	if name == "" {
		name = defaultSummary
	}
	return name
}

func scheduledEventExists(existing []discord.GuildScheduledEvent, event Event) bool {
	name := scheduledEventName(event)
	start := event.Start.UTC().Truncate(time.Minute)
	for _, scheduled := range existing {
		if scheduled.Name != name {
			continue
		}
		if scheduled.ScheduledStartTime.UTC().Truncate(time.Minute).Equal(start) {
			return true
		}
	}
	return false
}
