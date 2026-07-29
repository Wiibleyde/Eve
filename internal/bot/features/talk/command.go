package talk

import (
	"strings"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const CommandName = "talk"

const (
	OptionMessage = "message"
	OptionUser    = "mp"
)

const (
	msgNotOwner       = "Cette commande est réservée au propriétaire du bot."
	msgEmpty          = "Le message ne peut pas être vide."
	msgSent           = "Message envoyé."
	msgChannelFail    = "Impossible d'envoyer le message dans ce salon."
	msgDMClosed       = "Impossible d'envoyer un message privé à cet utilisateur : ses messages privés sont probablement fermés."
	msgUserUnresolved = "Impossible de déterminer le destinataire du message privé."
)

var messageMaxLength = 2000

var command = discord.SlashCommandCreate{
	Name:        CommandName,
	Description: "Faire parler le bot",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        OptionMessage,
			Description: "Le message à envoyer",
			Required:    true,
			MaxLength:   &messageMaxLength,
		},
		discord.ApplicationCommandOptionUser{
			Name:        OptionUser,
			Description: "Envoyer le message en privé à cet utilisateur au lieu du salon",
			Required:    false,
		},
	},
}

var DMSentHook func(target discord.User, content string)

func Commands() []discord.ApplicationCommandCreate {
	if !helpers.OwnerConfigured() {
		return nil
	}
	return []discord.ApplicationCommandCreate{command}
}

func allowedMentions() *discord.AllowedMentions {
	return &discord.AllowedMentions{
		Parse: []discord.AllowedMentionType{discord.AllowedMentionTypeUsers},
	}
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	if !helpers.IsOwner(e.User().ID) {
		logger.Warn("Non-owner attempted /talk", "user", e.User().ID.String())
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(msgNotOwner))
		return
	}

	data := e.SlashCommandInteractionData()
	message := strings.TrimSpace(data.String(OptionMessage))
	if message == "" {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(msgEmpty))
		return
	}

	if _, present := data.Option(OptionUser); present {
		target, ok := data.OptUser(OptionUser)
		if !ok {
			logger.Error("Error resolving /talk mp option target", "user", e.User().ID.String())
			helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(msgUserUnresolved))
			return
		}
		sendDM(e, target, message)
		return
	}
	sendToChannel(e, message)
}

func sendToChannel(e *events.ApplicationCommandInteractionCreate, message string) {
	channelID := e.Channel().ID()

	if _, err := e.Client().Rest.CreateMessage(channelID, discord.MessageCreate{
		Content:         message,
		AllowedMentions: allowedMentions(),
	}); err != nil {
		logger.Error("Error sending /talk message to channel",
			"channel", channelID.String(),
			"error", err,
		)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(msgChannelFail))
		return
	}

	logger.Info("Talk message sent", "channel", channelID.String(), "by", e.User().ID.String())
	helpers.RespondEphemeralEmbed(e, embeds.SuccessEmbed(msgSent))
}

func sendDM(e *events.ApplicationCommandInteractionCreate, target discord.User, message string) {
	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("Error deferring /talk DM interaction", "error", err)
		return
	}

	channel, err := e.Client().Rest.CreateDMChannel(target.ID)
	if err != nil {
		logger.Warn("Error opening DM channel for /talk", "target", target.ID.String(), "error", err)
		followupEphemeralEmbed(e, embeds.ErrorEmbed(msgDMClosed))
		return
	}

	if _, err := e.Client().Rest.CreateMessage(channel.ID(), discord.MessageCreate{
		Content:         message,
		AllowedMentions: allowedMentions(),
	}); err != nil {
		logger.Warn("Error sending /talk DM", "target", target.ID.String(), "error", err)
		followupEphemeralEmbed(e, embeds.ErrorEmbed(msgDMClosed))
		return
	}

	logger.Info("Talk DM sent", "target", target.ID.String(), "by", e.User().ID.String())
	followupEphemeralEmbed(e, embeds.SuccessEmbed("Message privé envoyé à <@"+target.ID.String()+">."))

	if hook := DMSentHook; hook != nil {
		hook(target, message)
	}
}

func followupEphemeralEmbed(e *events.ApplicationCommandInteractionCreate, embed discord.Embed) {
	if _, err := e.Client().Rest.CreateFollowupMessage(e.ApplicationID(), e.Token(), discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	}); err != nil {
		logger.Error("Error sending /talk DM followup", "error", err)
	}
}
