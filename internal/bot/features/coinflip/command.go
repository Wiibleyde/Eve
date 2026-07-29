// Package coinflip implements the /coinflip slash command: a public coin toss.
//
// The reply is public (not ephemeral) on purpose — the point of a coin flip is
// that everyone in the channel sees the same result.
//
// Randomness comes from math/rand/v2's global source, which is seeded from the
// runtime's entropy pool at startup and is safe for concurrent use, so no
// package-level generator or mutex is needed.
package coinflip

import (
	"math/rand/v2"

	"Eve/internal/bot/embeds"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

// CommandName is the slash command name, also used as the router key.
const CommandName = "coinflip"

// EdgeOdds is the 1-in-N chance of the coin landing on its edge. It is a pure
// flourish: it never skews the Pile/Face split, which stays 50/50 among the
// remaining outcomes.
const EdgeOdds = 6000

// User-facing strings (French, per project conventions).
const (
	embedTitle = "Pile ou face"

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
	embed := embeds.BaseEmbed()
	embed.Title = embedTitle
	embed.Description = flip()

	if err := e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	}); err != nil {
		logger.Error("Error responding to /coinflip", "error", err)
	}
}
