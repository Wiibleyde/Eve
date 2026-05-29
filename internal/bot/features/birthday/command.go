package birthday

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/database"
	"Eve/internal/database/ent/birthday"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "birthday",
		Description: "Gestion des anniversaire",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "set",
				Description: "Ajouter votre anniversaire",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{Name: "date", Description: "Date au format DD/MM/YYYY", Required: true},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "remove",
				Description: "Retirer votre anniversaire",
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "list",
				Description: "Lister les anniversaires du serveur",
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "adminset",
				Description: "[ADMIN] Définir la date d'un utilisateur",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionUser{Name: "user", Description: "Utilisateur", Required: true},
					discord.ApplicationCommandOptionString{Name: "date", Description: "Date au format DD/MM/YYYY", Required: true},
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
	case "set":
		handleSetBirthday(e)
	case "remove":
		handleRemoveBirthday(e)
	case "list":
		handleListBirthdays(e)
	case "adminset":
		handleAdminSetBirthday(e)
	}
}

func handleSetBirthday(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()
	dateStr := data.String("date")

	date, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Format invalide. Utilisez DD/MM/YYYY."))
		return
	}

	ctx := context.Background()
	discordID := e.User().ID.String()

	n, err := database.Default.Ent().Birthday.Update().
		Where(birthday.DiscordID(discordID)).
		SetBirthday(date).
		Save(ctx)
	if err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement."))
		return
	}

	if n == 0 {
		err = database.Default.Ent().Birthday.Create().
			SetDiscordID(discordID).
			SetBirthday(date).
			Exec(ctx)
		if err != nil {
			helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement."))
			return
		}
	}

	helpers.RespondEphemeralEmbed(e, embeds.SuccessEmbed("Anniversaire enregistré !"))
}

func handleRemoveBirthday(e *events.ApplicationCommandInteractionCreate) {
	ctx := context.Background()
	discordID := e.User().ID.String()

	n, err := database.Default.Ent().Birthday.Delete().
		Where(birthday.DiscordID(discordID)).
		Exec(ctx)
	if err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de la suppression."))
		return
	}
	if n == 0 {
		helpers.RespondEphemeral(e, "Aucun anniversaire enregistré.")
		return
	}

	helpers.RespondEphemeralEmbed(e, embeds.SuccessEmbed("Anniversaire supprimé."))
}

func handleListBirthdays(e *events.ApplicationCommandInteractionCreate) {
	ctx := context.Background()

	birthdays, err := database.Default.Ent().Birthday.Query().All(ctx)
	if err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de la récupération des anniversaires."))
		return
	}
	if len(birthdays) == 0 {
		helpers.RespondEphemeral(e, "Aucun anniversaire enregistré sur ce serveur.")
		return
	}

	lines := make([]string, 0, len(birthdays))
	for _, b := range birthdays {
		lines = append(lines, fmt.Sprintf("<@%s> — %s", b.DiscordID, b.Birthday.Format("02/01")))
	}

	embed := embeds.BaseEmbed()
	embed.Title = "Anniversaires"
	embed.Description = strings.Join(lines, "\n")
	helpers.RespondEphemeralEmbed(e, embed)
}

func handleAdminSetBirthday(e *events.ApplicationCommandInteractionCreate) {
	if !helpers.RequirePermission(e, discord.PermissionAdministrator, "Vous n'avez pas la permission d'utiliser cette commande.") {
		return
	}

	data := e.SlashCommandInteractionData()
	target := data.User("user")
	dateStr := data.String("date")

	date, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Format invalide. Utilisez DD/MM/YYYY."))
		return
	}

	ctx := context.Background()
	discordID := target.ID.String()

	n, err := database.Default.Ent().Birthday.Update().
		Where(birthday.DiscordID(discordID)).
		SetBirthday(date).
		Save(ctx)
	if err != nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement."))
		return
	}

	if n == 0 {
		err = database.Default.Ent().Birthday.Create().
			SetDiscordID(discordID).
			SetBirthday(date).
			Exec(ctx)
		if err != nil {
			helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'enregistrement."))
			return
		}
	}

	helpers.RespondEphemeralEmbed(e, embeds.SuccessEmbed(fmt.Sprintf("Anniversaire de <@%s> enregistré !", discordID)))
}
