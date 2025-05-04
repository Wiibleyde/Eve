package events

// StreamData represents a Twitch stream
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

// UserData represents a Twitch user
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

// StreamEventHandler interface for handling stream events
type StreamEventHandler interface {
	OnNewStream(stream StreamData, userData UserData)
	OnStreamEnd(stream StreamData, userData UserData)
}

var streamEventHandlers []StreamEventHandler

// RegisterStreamEventHandler adds a handler for stream events
func RegisterStreamEventHandler(handler StreamEventHandler) {
	streamEventHandlers = append(streamEventHandlers, handler)
}

// NotifyNewStream calls all registered handlers for new stream events
func NotifyNewStream(stream StreamData, userData UserData) {
	for _, handler := range streamEventHandlers {
		handler.OnNewStream(stream, userData)
	}
}

// NotifyStreamEnd calls all registered handlers for stream end events
func NotifyStreamEnd(stream StreamData, userData UserData) {
	for _, handler := range streamEventHandlers {
		handler.OnStreamEnd(stream, userData)
	}
}
