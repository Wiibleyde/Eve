package intelligence

import (
	"context"
	"main/pkg/config"

	"google.golang.org/genai"
)

var (
	ai       *genai.Client
	aiInited bool = false
)

func InitAi() {
	// Create a new context
	ctx := context.Background()

	// Initialize the AI client with the default configuration
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: config.GetConfig().GoogleApiKey,
	})
	if err != nil {
		panic(err)
	}
	ai = client
	aiInited = true
}
