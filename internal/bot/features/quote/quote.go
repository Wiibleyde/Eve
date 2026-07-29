package quote

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/database"
	"Eve/internal/database/ent/guildconfig"
	"Eve/internal/database/tables"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
)

const maxContentLength = 2000

const (
	MsgImageError   = "Erreur lors de la génération de l'image."
	MsgPersistError = "Erreur lors de l'enregistrement de la citation."
	MsgSendError    = "Erreur lors de l'envoi de la citation."
	MsgTooLong      = "Cette citation est trop longue pour être publiée (2000 caractères maximum)."
	MsgNoText       = "Ce message ne contient pas de texte."
	MsgGuildOnly    = "Cette commande doit être utilisée dans un serveur."
)

var (
	ErrTooLong = errors.New("quote: rendered content exceeds the Discord limit")
	ErrImage   = errors.New("quote: image generation failed")
	ErrPersist = errors.New("quote: storing the quote failed")
)

type Result struct {
	Message   discord.MessageCreate
	ChannelID snowflake.ID
	Posted    bool
}

func Create(
	ctx context.Context,
	client *bot.Client,
	guildID snowflake.ID,
	author discord.User,
	text string,
	quoteContext string,
) (Result, error) {
	content := formatContent(text, author.ID, quoteContext)
	if utf8.RuneCountInString(content) > maxContentLength {
		return Result{}, ErrTooLong
	}

	imgData, err := generateImage(text, author.Username, quoteContext, author.EffectiveAvatarURL())
	if err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrImage, err)
	}

	var ctxPtr *string
	if quoteContext != "" {
		ctxPtr = &quoteContext
	}

	err = database.Default.Ent().Quote.Create().
		SetID(uuid.New().String()).
		SetGuildID(guildID.String()).
		SetAuthorID(author.ID.String()).
		SetQuote(text).
		SetNillableContext(ctxPtr).
		Exec(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrPersist, err)
	}

	msg := discord.MessageCreate{
		Content: content,
		Files: []*discord.File{
			discord.NewFile("quote.png", "", bytes.NewReader(imgData)),
		},
	}

	channelID, configured := configuredChannel(ctx, guildID)
	if !configured {
		return Result{Message: msg}, nil
	}

	if _, err := client.Rest.CreateMessage(channelID, msg); err != nil {
		logger.Error("sending quote to channel: " + err.Error())
	}
	return Result{Message: msg, ChannelID: channelID, Posted: true}, nil
}

func ErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrTooLong):
		return MsgTooLong
	case errors.Is(err, ErrImage):
		return MsgImageError
	case errors.Is(err, ErrPersist):
		return MsgPersistError
	default:
		return MsgSendError
	}
}

func formatContent(text string, authorID snowflake.ID, quoteContext string) string {
	content := "\"" + text + "\" <@" + authorID.String() + "> - " + time.Now().Format("02/01/2006")
	if quoteContext != "" {
		content += " - *" + quoteContext + "*"
	}
	return content
}

func configuredChannel(ctx context.Context, guildID snowflake.ID) (snowflake.ID, bool) {
	cfg, err := database.Default.Ent().GuildConfig.Query().
		Where(
			guildconfig.GuildID(guildID.String()),
			guildconfig.Key(tables.QuoteChannel.String()),
		).
		Only(ctx)
	if err != nil {
		return 0, false
	}

	channelID, err := snowflake.Parse(cfg.Value)
	if err != nil {
		logger.Error("parsing quote channel ID: " + err.Error())
		return 0, false
	}
	return channelID, true
}

func respond(e *events.ApplicationCommandInteractionCreate, res Result) {
	if res.Posted {
		helpers.RespondEphemeralEmbed(e, embeds.SuccessEmbed("Citation envoyée dans <#"+res.ChannelID.String()+">."))
		return
	}
	if err := e.CreateMessage(res.Message); err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(MsgSendError))
	}
}

func guildID(e *events.ApplicationCommandInteractionCreate) (snowflake.ID, bool) {
	id := e.GuildID()
	if id == nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(MsgGuildOnly))
		return 0, false
	}
	return *id, true
}
