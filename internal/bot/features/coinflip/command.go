package coinflip

import (
	"math/rand/v2"

	"Eve/internal/bot/ui"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const CommandName = "coinflip"

const EdgeOdds = 6000

const (
	cardTitle = "🪙 Pile ou face"

	msgHeads = "Le résultat du lancer de pièce est : **Pile**."
	msgTails = "Le résultat du lancer de pièce est : **Face**."
	msgEdge  = "La pièce est tombée sur la tranche. 🪙"
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        CommandName,
		Description: "Lancer une pièce",
	},
}

func flip() string {
	if rand.IntN(EdgeOdds) == 0 {
		return msgEdge
	}
	if rand.IntN(2) == 0 {
		return msgHeads
	}
	return msgTails
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	card := ui.New().Title(cardTitle).Text(flip())

	if err := e.CreateMessage(card.MessageCreate()); err != nil {
		logger.Error("Error responding to /coinflip", "error", err)
	}
}
