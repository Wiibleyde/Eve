package bot

import (
	"main/pkg/modalHandler"

	"github.com/bwmarrin/discordgo"
)

var (
	modalHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
		"motusTryModal": modalHandler.MotusTryModal,
	}
)
