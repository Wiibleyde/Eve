package twitch

import (
	"encoding/json"
	"main/pkg/config"
	"main/pkg/data"
	"main/pkg/logger"
	"net/http"
	"strings"
	"time"
)

const (
	OnlineUrl  = "https://api.twitch.tv/helix/streams"
	OfflineUrl = "https://api.twitch.tv/helix/users"
)

type StreamData struct {
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
}

type OnlineStreamResponse struct {
	Data       []StreamData `json:"data"`
	Pagination struct {
		Cursor string `json:"cursor"`
	} `json:"pagination"`
}

type UserData struct {
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
}

type OfflineStreamResponse struct {
	Data []UserData `json:"data"`
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
	oauthToken := GetOauthToken()
	header := http.Header{}
	header.Add("Authorization", "Bearer "+oauthToken)
	header.Add("Client-Id", config.GetConfig().TwitchClientId)
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

// Map pour stocker l'état actuel des streams (true = en ligne, false = hors ligne)
var streamersState = make(map[string]StreamData)

func StartAutomaticStreamCheck() {
	for {
		onlineStreams, offlineStreams, err := GetStreamers()
		if err != nil {
			logger.ErrorLogger.Println("Error fetching streamers:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// Créer une map pour suivre les streamers actuellement en ligne
		currentOnlineStreamers := make(map[string]StreamData)

		// Vérifier les nouveaux streams et mettre à jour la map des streamers en ligne
		for _, stream := range onlineStreams.Data {
			currentOnlineStreamers[stream.UserID] = stream

			// Vérifier si le streamer n'était pas en ligne avant
			if _, exists := streamersState[stream.UserID]; !exists {
				// Trouver les données utilisateur correspondantes
				var userData UserData
				for _, user := range offlineStreams.Data {
					if user.ID == stream.UserID {
						userData = user
						break
					}
				}
				// Appeler l'événement de nouveau stream
				OnNewStream(stream, userData)
			}
		}

		// Vérifier les streams qui se sont terminés
		for userID, stream := range streamersState {
			if _, stillOnline := currentOnlineStreamers[userID]; !stillOnline {
				// Trouver les données utilisateur correspondantes
				var userData UserData
				for _, user := range offlineStreams.Data {
					if user.ID == userID {
						userData = user
						break
					}
				}
				// Appeler l'événement de fin de stream
				OnStreamEnd(stream, userData)
			}
		}

		// Mettre à jour la map d'état global
		streamersState = currentOnlineStreamers

		time.Sleep(10 * time.Second) // Vérification toutes les 10 secondes
	}
}

// Event on new stream
func OnNewStream(stream StreamData, userData UserData) {
	// Handle new stream event
	logger.DebugLogger.Printf("New stream started: %s is now live playing %s\n", stream.UserName, stream.GameName)
}

// Event on stream end
func OnStreamEnd(stream StreamData, userData UserData) {
	// Handle stream end event
	logger.DebugLogger.Printf("Stream ended: %s is now offline\n", stream.UserName)
}
