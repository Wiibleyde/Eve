package twitch

import (
	"encoding/json"
	"main/pkg/config"
	"main/pkg/logger"
	"net/http"
	"net/url"
	"time"
)

var (
	OauthToken string
	ExpiresAt  int64
)

const (
	OauthUrl  = "https://id.twitch.tv/oauth2/token"
	GrantType = "client_credentials"
)

type TwitchAuthToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func refreshToken() {
	params := url.Values{}
	params.Add("client_id", config.GetConfig().TwitchClientId)
	params.Add("client_secret", config.GetConfig().TwitchClientSecret)
	params.Add("grant_type", GrantType)

	resp, err := http.PostForm(OauthUrl, params)
	if err != nil {
		logger.ErrorLogger.Println("Error getting new Twitch token:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.ErrorLogger.Println("Error getting new Twitch token: status code", resp.StatusCode)
		return
	}
	var token TwitchAuthToken
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		logger.ErrorLogger.Println("Error decoding Twitch token:", err)
		return
	}
	OauthToken = token.AccessToken
	ExpiresAt = time.Now().Unix() + token.ExpiresIn
	logger.DebugLogger.Println("New Twitch token obtained:", OauthToken)
}

func isTokenExpired() bool {
	if OauthToken == "" {
		return true
	}
	if time.Now().Unix() >= ExpiresAt {
		return true
	}
	return false
}

// GetOauthToken returns the OAuth token, refreshing it if necessary and logging the process.
func GetOauthToken() string {
	if isTokenExpired() {
		refreshToken()
	}
	return OauthToken
}


