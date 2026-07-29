package quote

import (
	"context"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "quote",
		Description: "Gestion des citations",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "add",
				Description: "Ajouter une citation",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "quote",
						Description: "La citation",
						Required:    true,
					},
					discord.ApplicationCommandOptionUser{
						Name:        "author",
						Description: "L'auteur de la citation",
						Required:    true,
					},
					discord.ApplicationCommandOptionString{
						Name:        "context",
						Description: "Le contexte de la citation",
						Required:    false,
					},
				},
			},
		},
	},
	discord.MessageCommandCreate{
		Name: MenuCreateQuote,
	},
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()
	if data.SubCommandName == nil {
		return
	}
	switch *data.SubCommandName {
	case "add":
		handleAddQuote(e)
	}
}

func handleAddQuote(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()

	guild, ok := guildID(e)
	if !ok {
		return
	}

	quoteText := data.String("quote")
	author := data.User("author")
	quoteContext, _ := data.OptString("context")

	res, err := Create(context.Background(), e.Client(), guild, author, quoteText, quoteContext)
	if err != nil {
		logger.Error("creating quote", "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(ErrorMessage(err)))
		return
	}
	respond(e, res)
}
