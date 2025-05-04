package commandHandler

import (
	"main/pkg/bot_utils"
	"main/pkg/data"
	"main/pkg/logger"
	"main/prisma/db"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func BirthdayHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	subCommand := i.ApplicationCommandData().Options[0]
	switch subCommand.Name {
	case "set":
		if len(subCommand.Options) == 0 {
			return nil
		}
		strBirthDate := subCommand.Options[0].StringValue()

		// Try parsing with different formats
		var birthDate time.Time
		var err error

		// First try RFC3339 format (2006-01-02T15:04:05Z07:00)
		birthDate, err = time.Parse(time.RFC3339, strBirthDate)
		if err != nil {
			// Try simple date format (2006-01-02)
			birthDate, err = time.Parse("2006-01-02", strBirthDate)
			if err != nil {
				// Try other common formats
				formats := []string{"02/01/2006", "01/02/2006", "2006/01/02"}
				parsed := false

				for _, format := range formats {
					birthDate, err = time.Parse(format, strBirthDate)
					if err == nil {
						parsed = true
						break
					}
				}

				if !parsed {
					logger.ErrorLogger.Println("Error parsing birth date:", err)
					embed := bot_utils.ErrorEmbed(s, err.Error())
					embed.Description = "Format de date invalide. Veuillez utiliser un format de date valide (YYYY-MM-DD)"
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{embed},
							Flags:  discordgo.MessageFlagsEphemeral,
						},
					})
					return err
				}
			}
		}

		userId := i.Member.User.ID
		client, ctx := data.GetDBClient()
		_, err = client.GlobalUserData.UpsertOne(
			db.GlobalUserData.UserID.Equals(userId),
		).Create(
			db.GlobalUserData.UserID.Set(userId),
			db.GlobalUserData.BirthDate.Set(birthDate),
		).Update(
			db.GlobalUserData.BirthDate.Set(birthDate),
		).Exec(ctx)
		if err != nil {
			return err
		}
		embed := bot_utils.SuccessEmbed(s, "Date de naissance définie avec succès.")
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Date de naissance",
			Value:  birthDate.Format("02/01/2006"),
			Inline: true,
		})
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return err
		}
	case "get":
		userId := i.Member.User.ID

		// Check if a user ID was provided as an option
		if len(subCommand.Options) > 0 {
			userOption := subCommand.Options[0].UserValue(s)
			if userOption != nil {
				userId = userOption.ID
			}
		}

		client, ctx := data.GetDBClient()
		globalUserData, err := client.GlobalUserData.FindUnique(
			db.GlobalUserData.UserID.Equals(userId),
		).Exec(ctx)
		if err != nil {
			if err.Error() == "ErrNotFound" {
				embed := bot_utils.WarningEmbed(s, "Date de naissance demandée.")
				embed.Description = "Aucune date de naissance trouvée pour <@" + userId + ">."
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
			return err
		}
		birthDate := globalUserData.BirthDate
		embed := bot_utils.BasicEmbedBuilder(s)
		birthDateValue, _ := birthDate()
		if birthDateValue.IsZero() {
			embed = bot_utils.WarningEmbed(s, "Aucune date de naissance trouvée.")
			embed.Description = "Vous n'avez pas encore défini de date de naissance. Utilisez `/birthday set` pour le faire."
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
		embed.Color = 0xCC9845
		embed.Title = "Date de naissance demandée"
		embed.Description = "Voici la date de naissance de <@" + userId + "> : " + birthDateValue.Format("02/01/2006")
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return err
		}
	case "delete":
		userId := i.Member.User.ID
		client, ctx := data.GetDBClient()
		_, err := client.GlobalUserData.FindUnique(
			db.GlobalUserData.UserID.Equals(userId),
		).Update(
			db.GlobalUserData.BirthDate.SetOptional(nil),
		).Exec(ctx)
		if err != nil {
			return err
		}
		embed := bot_utils.SuccessEmbed(s, "Date de naissance supprimée avec succès.")
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return err
		}
	case "list":
		// Fetch members
		members, err := s.GuildMembers(i.GuildID, "", 1000)
		if err != nil {
			return err
		}

		// Extract user IDs from member objects
		userIDs := make([]string, len(members))
		for i, member := range members {
			userIDs[i] = member.User.ID
		}
		client, ctx := data.GetDBClient()
		birthdays, err := client.GlobalUserData.FindMany(
			db.GlobalUserData.UserID.In(userIDs),
		).OrderBy(
			db.GlobalUserData.BirthDate.Order(db.SortOrderAsc),
		).Select(
			db.GlobalUserData.UserID.Field(),
			db.GlobalUserData.BirthDate.Field(),
		).Exec(ctx)
		if err != nil {
			return err
		}
		if len(birthdays) == 0 {
			embed := bot_utils.WarningEmbed(s, "Aucune date de naissance trouvée.")
			embed.Description = "Aucune date de naissance trouvée dans le serveur."
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
		// Filter the birthdays IsZero
		var filteredBirthdays []db.GlobalUserDataModel
		for _, birthday := range birthdays {
			birthDateValue, _ := birthday.BirthDate()
			if !birthDateValue.IsZero() {
				filteredBirthdays = append(filteredBirthdays, birthday)
			}
		}
		const months = "Janvier Février Mars Avril Mai Juin Juillet Août Septembre Octobre Novembre Décembre"
		// Split the months string into a slice
		monthsList := strings.Split(months, " ")
		// Create a map to store birthdays by month
		birthdaysByMonth := make(map[string][]string)
		for _, birthday := range filteredBirthdays {
			birthDateValue, _ := birthday.BirthDate()
			month := monthsList[birthDateValue.Month()-1]
			birthdaysByMonth[month] = append(birthdaysByMonth[month], birthday.UserID)
		}
		// Create an embed for the response
		embed := bot_utils.BasicEmbedBuilder(s)
		embed.Title = "Dates de naissance sur le serveur"
		embed.Description = "Voici les dates de naissance des membres du serveur :"
		for month, userIDs := range birthdaysByMonth {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   month,
				Value:  "Membres : " + bot_utils.FormatUserIDs(userIDs),
				Inline: true,
			})
		}
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
