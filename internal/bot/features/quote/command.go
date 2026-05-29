package quote

import (
	"bytes"
	"context"
	"time"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/database"
	"Eve/internal/database/ent/guildconfig"
	"Eve/internal/database/tables"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
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

	quoteText := data.String("quote")
	author := data.User("author")
	ctxVal, hasCtx := data.OptString("context")

	guildID := e.GuildID()
	if guildID == nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Cette commande doit être utilisée dans un serveur."))
		return
	}

	var ctxPtr *string
	contextStr := ""
	if hasCtx {
		ctxPtr = &ctxVal
		contextStr = ctxVal
	}

	imgData, err := generateImage(quoteText, author.Username, contextStr, author.EffectiveAvatarURL())
	if err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de la génération de l'image."))
		return
	}

	dbCtx := context.Background()
	err = database.Default.Ent().Quote.Create().
		SetID(uuid.New().String()).
		SetGuildID(guildID.String()).
		SetAuthorID(author.ID.String()).
		SetQuote(quoteText).
		SetNillableContext(ctxPtr).
		Exec(dbCtx)
	if err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement de la citation."))
		return
	}

	var content string
	content = "\"" + quoteText + "\" <@" + author.ID.String() + "> - " + time.Now().Format("02/01/2006")
	if len(contextStr) > 0 {
		content = content + " - *" + contextStr + "*"
	}

	msg := discord.MessageCreate{
		Content: content,
		Files: []*discord.File{
			discord.NewFile("quote.png", "", bytes.NewReader(imgData)),
		},
	}

	cfg, err := database.Default.Ent().GuildConfig.Query().
		Where(guildconfig.GuildID(guildID.String()), guildconfig.Key(tables.QuoteChannel.String())).
		Only(dbCtx)

	if err == nil {
		// Configured channel: send there and acknowledge the interaction ephemerally.
		channelID, parseErr := snowflake.Parse(cfg.Value)
		if parseErr != nil {
			logger.Error("parsing quote channel ID: " + parseErr.Error())
		} else {
			if _, sendErr := e.Client().Rest.CreateMessage(channelID, msg); sendErr != nil {
				logger.Error("sending quote to channel: " + sendErr.Error())
			}
			helpers.RespondEphemeralEmbed(e, embeds.SuccessEmbed("Citation envoyée dans <#"+cfg.Value+">."))
			return
		}
	}

	// No channel configured (or parse error): respond in current channel.
	if err := e.CreateMessage(msg); err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'envoi de la citation."))
	}
}
