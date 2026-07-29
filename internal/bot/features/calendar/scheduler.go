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
	// TickInterval is how often every configured calendar is refreshed.
	TickInterval = 5 * time.Minute
	// startupDelay lets the gateway settle before the first tick so the embed
	// is fresh shortly after a restart instead of one full interval later.
	startupDelay = 30 * time.Second
	// scheduledEventLead is how long before an event starts the matching
	// Discord scheduled event is created.
	scheduledEventLead = 30 * time.Minute
	// tickTimeout bounds one whole pass so a hanging calendar can never stall
	// the loop past the next tick.
	tickTimeout = TickInterval - 30*time.Second

	// Discord limits for scheduled events.
	maxScheduledEventName        = 100
	maxScheduledEventDescription = 1000
	maxScheduledEventLocation    = 100
	// defaultEventLocation satisfies the "location is required for external
	// events" rule when the VEVENT carries no LOCATION.
	defaultEventLocation = "Non précisé"
)

// StartScheduler runs the calendar refresh loop in its own goroutine.
//
// Each tick refreshes every guild that has a calendar.url, isolating failures
// per guild, and creates the Discord scheduled events for anything starting
// within the next 30 minutes. Ticks are skipped entirely during maintenance.
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

// refreshGuildSafe isolates one guild: a panic or an error there never aborts
// the loop for the other guilds.
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

	// A missing message ref is normal right after the refs were cleared; the
	// guild republishes with /calendar set or /calendar refresh.
	if err := updateCalendarMessage(ctx, client, cfg, buildEmbed(parsed, now)); err != nil &&
		!errors.Is(err, errCalendarMessageGone) && !errors.Is(err, errNoCalendarMessage) {
		logger.Warn("Calendar: refreshing message failed", "guild", guildIDStr, "error", err)
	}

	return syncScheduledEvents(client, guildID, parsed, now)
}

// syncScheduledEvents creates a Discord scheduled event for every calendar
// event starting within the lead window.
//
// Dedupe is done purely from the guild's current scheduled events (name +
// start time), which makes it restart-safe with no state of our own.
//
// Discord omits CANCELED and COMPLETED events from that listing, so an event a
// moderator cancels while its calendar entry is still inside the lead window is
// recreated on the next tick. Remembering cancellations would mean persisting
// every created event — not worth it for a 30 minute window.
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
		// Keep the local view in sync so two calendar entries with the same
		// name and start inside one tick do not both get created.
		existing = append(existing, *created)
		logger.Info("Calendar: scheduled event created",
			"guild", guildID.String(), "event", created.Name, "start", created.ScheduledStartTime.Format(time.RFC3339))
	}
	return nil
}

// eventsStartingWithin returns the events starting strictly after now and no
// later than now+lead. Already-started events are excluded: Discord rejects a
// scheduled event whose start is in the past.
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

// scheduledEventName is the single source of truth for the name used both when
// creating and when deduping, so truncation can never break the match.
func scheduledEventName(event Event) string {
	name := inline(event.Summary, maxScheduledEventName)
	if name == "" {
		name = defaultSummary
	}
	return name
}

// scheduledEventExists matches on name + start time (minute precision, since
// Discord echoes back timestamps it normalised).
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
