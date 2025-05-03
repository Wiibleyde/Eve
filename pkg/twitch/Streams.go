package twitch

import (
	"encoding/json"
	"main/pkg/data"
	"net/http"
	"strings"
	"time"
)

const (
	OnlineUrl  = "https://api.twitch.tv/helix/streams"
	OfflineUrl = "https://api.twitch.tv/helix/users"
)

type OnlineStreamResponse struct {
	Data []struct {
		ID           string   `json:"id"`
		UserID       string   `json:"user_id"`
		UserLogin    string   `json:"user_login"`
		UserName     string   `json:"user_name"`
		GameID       string   `json:"game_id"`
		GameName     string   `json:"game_name"`
		Type         string   `json:"type"`
		Title        string   `json:"title"`
		Tags         []string `json:"tags"`
		ViewerCount  int      `json:"viewer_count"`
		StartedAt    string   `json:"started_at"`
		Language     string   `json:"language"`
		ThumbnailURL string   `json:"thumbnail_url"`
		IsMature     bool     `json:"is_mature"`
	} `json:"data"`
	Pagination struct {
		Cursor string `json:"cursor"`
	} `json:"pagination"`
}

type OfflineStreamResponse struct {
	Data []struct {
		ID              string `json:"id"`
		Login           string `json:"login"`
		DisplayName     string `json:"display_name"`
		Type            string `json:"type"`
		BroadcasterType string `json:"broadcaster_type"`
		Description     string `json:"description"`
		ProfileImageURL string `json:"profile_image_url"`
		OfflineImageURL string `json:"offline_image_url"`
		ViewCount       int    `json:"view_count"`
		Email           string `json:"email"`
		CreatedAt       string `json:"created_at"`
	} `json:"data"`
}

var (
	OnlineStreams  OnlineStreamResponse
	OfflineStreams OfflineStreamResponse
)

func GenerateOnlineUrl() string {
	client, ctx := data.GetDBClient()
	streamers, err := client.Stream.FindMany().Exec(ctx)
	if err != nil {
		return ""
	}
	uniqueMap := make(map[string]struct{})
	var uniqueStreamers []string

	for _, streamer := range streamers {
		if _, exists := uniqueMap[streamer.TwitchChannelName]; !exists {
			uniqueMap[streamer.TwitchChannelName] = struct{}{}
			uniqueStreamers = append(uniqueStreamers, "user_login="+streamer.TwitchChannelName)
		}
	}

	// Limit to 100 streamers
	if len(uniqueStreamers) > 100 {
		uniqueStreamers = uniqueStreamers[:100]
	}

	return strings.Join(uniqueStreamers, "&")
}

func GenerateOfflineUrl() string {
	client, ctx := data.GetDBClient()
	streamers, err := client.Stream.FindMany().Exec(ctx)
	if err != nil {
		return ""
	}
	uniqueMap := make(map[string]struct{})
	var uniqueStreamers []string

	for _, streamer := range streamers {
		if _, exists := uniqueMap[streamer.TwitchChannelName]; !exists {
			uniqueMap[streamer.TwitchChannelName] = struct{}{}
			uniqueStreamers = append(uniqueStreamers, "user_id="+streamer.TwitchChannelName)
		}
	}

	// Limit to 100 streamers
	if len(uniqueStreamers) > 100 {
		uniqueStreamers = uniqueStreamers[:100]
	}

	return strings.Join(uniqueStreamers, "&")
}

func GetOnlineStreamers() (OnlineStreamResponse, error) {
	generateUrlParameters := GenerateOnlineUrl()
	if generateUrlParameters == "" {
		return OnlineStreamResponse{}, nil
	}
	url := OnlineUrl + "?" + generateUrlParameters
	response, err := http.Get(url)
	if err != nil {
		return OnlineStreamResponse{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return OnlineStreamResponse{}, nil
	}
	err = json.NewDecoder(response.Body).Decode(&OnlineStreams)
	if err != nil {
		return OnlineStreamResponse{}, err
	}
	return OnlineStreams, nil
}

func GetOfflineStreamers() (OfflineStreamResponse, error) {
	generateUrlParameters := GenerateOfflineUrl()
	if generateUrlParameters == "" {
		return OfflineStreamResponse{}, nil
	}
	url := OfflineUrl + "?" + generateUrlParameters
	response, err := http.Get(url)
	if err != nil {
		return OfflineStreamResponse{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return OfflineStreamResponse{}, nil
	}
	err = json.NewDecoder(response.Body).Decode(&OfflineStreams)
	if err != nil {
		return OfflineStreamResponse{}, err
	}
	return OfflineStreams, nil
}

func GetStreamers() (OnlineStreamResponse, OfflineStreamResponse, error) {
	onlineStreams, err := GetOnlineStreamers()
	if err != nil {
		return OnlineStreamResponse{}, OfflineStreamResponse{}, err
	}
	offlineStreams, err := GetOfflineStreamers()
	if err != nil {
		return OnlineStreamResponse{}, OfflineStreamResponse{}, err
	}
	return onlineStreams, offlineStreams, nil
}

func StartAutomaticStreamCheck() {
	for {
		onlineStreams, offlineStreams, err := GetStreamers()
		if err != nil {
			return
		}
		OnlineStreams = onlineStreams
		OfflineStreams = offlineStreams

		time.Sleep(7 * time.Second) // Adjust the sleep duration as needed
	}
}
