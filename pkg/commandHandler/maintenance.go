package commandHandler

import (
	"main/pkg/bot_utils"
	"main/pkg/config"

	"github.com/bwmarrin/discordgo"
)

func MaintenanceHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check if the user is an admin
	if i.Member.User.ID != config.GetConfig().OwnerId {
		embed := bot_utils.WarningEmbed(s, "Vous n'avez pas la permission d'utiliser cette commande.")
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return err
		}
		return nil
	}

	isMaintenance := bot_utils.IsMaintenanceMode()
	word := "désactivé"
	if isMaintenance {
		bot_utils.SetMaintenanceMode(false)
	} else {
		bot_utils.SetMaintenanceMode(true)
		word = "activé"
	}

	embed := bot_utils.SuccessEmbed(s, "Le mode maintenance a été "+word+".")
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
