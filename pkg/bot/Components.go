package bot

import (
	"main/pkg/buttonHandler"

	"github.com/bwmarrin/discordgo"
)

var (
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
		"jokeSetPublicButton": buttonHandler.JokeSetPublicButton,
		"motusTryButton":      buttonHandler.MotusTryButton,
	}
)
