package twitch

import (
	"encoding/json"
	"main/pkg/config"
	"main/pkg/data"
	"main/pkg/events"
	"main/pkg/logger"
	"net/http"
	"strings"
	"time"
)

const (
	OnlineUrl  = "https://api.twitch.tv/helix/streams"
	OfflineUrl = "https://api.twitch.tv/helix/users"
)

// Using the StreamData from events package
type StreamData = events.StreamData

type OnlineStreamResponse struct {
	Data       []StreamData `json:"data"`
	Pagination struct {
		Cursor string `json:"cursor"`
	} `json:"pagination"`
}

// Using the UserData from events package
type UserData = events.UserData

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
			uniqueStreamers = append(uniqueStreamers, "login="+streamer.TwitchChannelName)
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

	// Create a new request with the headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return OnlineStreamResponse{}, err
	}

	// Set the headers on the request
	req.Header = header

	// Make the request
	client := &http.Client{}
	response, err := client.Do(req)
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
	oauthToken := GetOauthToken()
	header := http.Header{}
	header.Add("Authorization", "Bearer "+oauthToken)
	header.Add("Client-Id", config.GetConfig().TwitchClientId)
	generateUrlParameters := GenerateOfflineUrl()
	if generateUrlParameters == "" {
		logger.ErrorLogger.Println("No streamers found")
		return OfflineStreamResponse{}, nil
	}
	url := OfflineUrl + "?" + generateUrlParameters

	// Create a new request with the headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.ErrorLogger.Println("Error creating request:", err)
		return OfflineStreamResponse{}, err
	}

	// Set the headers on the request
	req.Header = header

	// Make the request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		logger.ErrorLogger.Println("Error making request:", err)
		return OfflineStreamResponse{}, err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logger.ErrorLogger.Println("Error: received non-200 response code:", response.StatusCode)
		return OfflineStreamResponse{}, nil
	}
	err = json.NewDecoder(response.Body).Decode(&OfflineStreams)
	if err != nil {
		logger.ErrorLogger.Println("Error decoding response:", err)
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
	logger.DebugLogger.Printf("New stream started: %s is now live playing %s", stream.UserName, stream.GameName)
	// Use events package to notify handlers instead of directly calling bot function
	events.NotifyNewStream(stream, userData)
}

// Event on stream end
func OnStreamEnd(stream StreamData, userData UserData) {
	// Handle stream end event
	logger.DebugLogger.Printf("Stream ended: %s is now offline", stream.UserName)
	// Use events package to notify handlers instead of directly calling bot function
	events.NotifyStreamEnd(stream, userData)
}
