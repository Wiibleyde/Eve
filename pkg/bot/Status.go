package bot

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

type StatusConfig struct {
	Period struct {
		StartDate time.Time `json:"startDate"`
		EndDate   time.Time `json:"endDate"`
	} `json:"period"`
	Status []struct {
		Name string                 `json:"name"`
		Type discordgo.ActivityType `json:"type"`
	} `json:"status"`
}

var (
	halloweenStatus = StatusConfig{
		Period: struct {
			StartDate time.Time `json:"startDate"`
			EndDate   time.Time `json:"endDate"`
		}{
			StartDate: time.Date(2023, 10, 24, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		},
		Status: []struct {
			Name string                 `json:"name"`
			Type discordgo.ActivityType `json:"type"`
		}{
			{
				Name: "la préparation des citrouilles 🎃",
				Type: discordgo.ActivityTypeCompeting,
			},
			{
				Name: "les fantômes... 👻",
				Type: discordgo.ActivityTypeWatching,
			},
			{
				Name: "Spooky Scary Skeletons",
				Type: discordgo.ActivityTypeListening,
			},
			{
				Name: "les bonbons ou un sort ! 🍬",
				Type: discordgo.ActivityTypeGame,
			},
		},
	}

	christmasStatus = StatusConfig{
		Period: struct {
			StartDate time.Time `json:"startDate"`
			EndDate   time.Time `json:"endDate"`
		}{
			StartDate: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 12, 26, 0, 0, 0, 0, time.UTC),
		},
		Status: []struct {
			Name string                 `json:"name"`
			Type discordgo.ActivityType `json:"type"`
		}{
			{
				Name: "l'emballage des cadeaux 🎁",
				Type: discordgo.ActivityTypeCompeting,
			},
			{
				Name: "les lutins 🧝",
				Type: discordgo.ActivityTypeWatching,
			},
			{
				Name: "les chants de Noël",
				Type: discordgo.ActivityTypeListening,
			},
			{
				Name: "le Père Noël 🎅",
				Type: discordgo.ActivityTypeGame,
			},
		},
	}

	defaultStatus = StatusConfig{
		Period: struct {
			StartDate time.Time `json:"startDate"`
			EndDate   time.Time `json:"endDate"`
		}{
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		Status: []struct {
			Name string                 `json:"name"`
			Type discordgo.ActivityType `json:"type"`
		}{
			{
				Name: "les merveilles de ce monde.",
				Type: discordgo.ActivityTypeWatching,
			},
			{
				Name: "vos instructions.",
				Type: discordgo.ActivityTypeListening,
			},
			{
				Name: "les données de mission.",
				Type: discordgo.ActivityTypeWatching,
			},
			{
				Name: "les étoiles.",
				Type: discordgo.ActivityTypeWatching,
			},
		},
	}
)

func startStatusChange(s *discordgo.Session) {
	currentIndex := 0
	var currentConfig StatusConfig

	for {
		currentTime := time.Now().UTC()

		if currentTime.After(halloweenStatus.Period.StartDate) && currentTime.Before(halloweenStatus.Period.EndDate) {
			currentConfig = halloweenStatus
		} else if currentTime.After(christmasStatus.Period.StartDate) && currentTime.Before(christmasStatus.Period.EndDate) {
			currentConfig = christmasStatus
		} else {
			currentConfig = defaultStatus
		}

		if len(currentConfig.Status) > 0 {
			setStatus(s, currentConfig, currentIndex%len(currentConfig.Status))

			currentIndex++
		}

		time.Sleep(10 * time.Second)
	}
}

func setStatus(s *discordgo.Session, statusConfig StatusConfig, index int) {
	// Set the status
	s.UpdateStatusComplex(discordgo.UpdateStatusData{
		IdleSince: nil,
		Activities: []*discordgo.Activity{
			{
				Name: statusConfig.Status[index].Name,
				Type: statusConfig.Status[index].Type,
			},
		},
		Status: "online",
		AFK:    false,
	})
}
