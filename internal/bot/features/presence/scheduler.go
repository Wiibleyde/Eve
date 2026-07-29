package presence

import (
	"context"
	"sync"
	"time"

	"Eve/internal/bot/maintenance"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
)

const TickInterval = 30 * time.Second

type activityKind int

const (
	watching activityKind = iota
	listening
	playing
	competing
)

type state struct {
	kind   activityKind
	name   string
	status discord.OnlineStatus
}

type activity struct {
	kind activityKind
	name string
}

var maintenanceActivity = activity{watching, "Regarde la maintenance"}

var defaultActivities = []activity{
	{watching, "Regarde les merveilles de ce monde."},
	{listening, "Écoute vos instructions."},
	{watching, "Regarde les données de mission."},
	{watching, "Regarde les étoiles."},
}

var halloweenActivities = []activity{
	{competing, "Participe à la préparation des citrouilles. 🎃"},
	{watching, "Regarde les fantômes... 👻"},
	{listening, "Spooky Scary Skeletons"},
	{playing, "Joue à des bonbons ou un sort ! 🍬"},
}

var christmasActivities = []activity{
	{competing, "Participe à l'emballage des cadeaux. 🎁"},
	{watching, "Regarde les lutins. 🧝"},
	{listening, "Chante des chants de Noël"},
	{playing, "Joue avec le Père Noël. 🎅"},
}

const (
	listDefault   = "default"
	listHalloween = "halloween"
	listChristmas = "christmas"
)

var schedulerOnce sync.Once

func StartScheduler(client *bot.Client) {
	schedulerOnce.Do(func() {
		go func() {
			r := &rotator{indices: make(map[string]int)}
			r.tick(client)

			ticker := time.NewTicker(TickInterval)
			defer ticker.Stop()
			for range ticker.C {
				r.tick(client)
			}
		}()
	})
}

type rotator struct {
	indices  map[string]int
	last     state
	lastSent bool
}

func (r *rotator) tick(client *bot.Client) {
	next := r.compute(time.Now())
	if r.lastSent && next == r.last {
		return
	}
	if err := apply(client, next); err != nil {
		logger.Error("Error updating presence", "error", err)
		return
	}
	r.last = next
	r.lastSent = true
	logger.Debug("Presence updated", "activity", next.name, "status", string(next.status))
}

func (r *rotator) compute(now time.Time) state {
	if maintenance.Enabled() {
		return state{
			kind:   maintenanceActivity.kind,
			name:   maintenanceActivity.name,
			status: discord.OnlineStatusDND,
		}
	}

	name, activities := activeList(now)
	a := activities[r.indices[name]%len(activities)]
	r.indices[name] = (r.indices[name] + 1) % len(activities)

	return state{kind: a.kind, name: a.name, status: discord.OnlineStatusOnline}
}

func activeList(now time.Time) (string, []activity) {
	switch {
	case inPeriod(now, time.October, 24, time.November, 7):
		return listHalloween, halloweenActivities
	case inPeriod(now, time.December, 1, time.December, 25):
		return listChristmas, christmasActivities
	default:
		return listDefault, defaultActivities
	}
}

func inPeriod(now time.Time, startMonth time.Month, startDay int, endMonth time.Month, endDay int) bool {
	year, loc := now.Year(), now.Location()
	start := time.Date(year, startMonth, startDay, 0, 0, 0, 0, loc)
	end := time.Date(year, endMonth, endDay+1, 0, 0, 0, 0, loc)
	return !now.Before(start) && now.Before(end)
}

func apply(client *bot.Client, s state) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.SetPresence(ctx,
		activityOpt(s.kind, s.name),
		gateway.WithOnlineStatus(s.status),
	)
}

func activityOpt(kind activityKind, name string) gateway.PresenceOpt {
	switch kind {
	case listening:
		return gateway.WithListeningActivity(name)
	case playing:
		return gateway.WithPlayingActivity(name)
	case competing:
		return gateway.WithCompetingActivity(name)
	default:
		return gateway.WithWatchingActivity(name)
	}
}
