package bot

import (
	"main/pkg/commandHandler"
	"main/pkg/contextMenuHandler"
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

var (
	commands = []*discordgo.ApplicationCommand{
		// Slash commands
		{
			Name:        "ping",
			Description: "Savoir si le bot est en ligne",
			// Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "birthday",
			Description: "Gérer les anniversaires",
			// Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "set",
					Description: "Définir votre date de naissance",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "date",
							Description: "Date de naissance au format YYYY-MM-DD",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
						},
					},
				},
				{
					Name:        "get",
					Description: "Obtenir votre date de naissance",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "user",
							Description: "Utilisateur dont vous voulez connaître la date de naissance",
							Type:        discordgo.ApplicationCommandOptionUser,
							Required:    false,
						},
					},
				},
				{
					Name:        "delete",
					Description: "Supprimer votre date de naissance",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "list",
					Description: "Lister les anniversaires (dans le serveur)",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
		{
			Name:        "blague",
			Description: "Obtenir une blague",
			// Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "type",
					Description: "Type de blague (certains types ont été censurés)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Globale",
							Value: "global",
						},
						{
							Name:  "Programmation",
							Value: "dev",
						},
						{
							Name:  "Beauf",
							Value: "beauf",
						},
						// {
						// 	Name:  "Dark",
						// 	Value: "dark",
						// },
						// {
						// 	Name:  "Limite",
						// 	Value: "limit",
						// },
						// {
						// 	Name:  "Blondes",
						// 	Value: "blondes",
						// },
					},
				},
			},
		},

		// Context menu user commands
		{
			Name: "Récupérer la photo de profil",
			Type: discordgo.UserApplicationCommand,
		},
		{
			Name: "Récupérer la bannière",
			Type: discordgo.UserApplicationCommand,
		},

		// Context menu message commands
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
		// Slash commands
		"ping":     commandHandler.PingHandler,
		"birthday": commandHandler.BirthdayHandler,
		"blague":   commandHandler.BlagueHandler,

		// Context menu user commands
		"Récupérer la photo de profil": contextMenuHandler.ProfilePictureContextMenuHandler,
		"Récupérer la bannière":        contextMenuHandler.BannerContextMenuHandler,
	}
)

// checkCommandHandlers check if every command has a handler
func checkCommandHandlers() {
	for i, v := range commands {
		if _, ok := commandHandlers[v.Name]; !ok {
			commands = append(commands[:i], commands[i+1:]...)
			logger.WarningLogger.Println("Removed command", v.Name)
			continue
		}
	}
}

func registerCommands(s *discordgo.Session) {
	checkCommandHandlers()

	// Remove previous commands
	_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", commands)
	if err != nil {
		logger.ErrorLogger.Printf("Impossible de supprimer les commandes précédentes: %v", err)
		return
	}
	logger.InfoLogger.Println("Suppression des commandes précédentes terminée !")

	// Register new commands
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))

	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			logger.ErrorLogger.Printf("Impossible d'enregistrer la commande '%v': %v", v.Name, err)
			continue
		}
		registeredCommands[i] = cmd
		logger.InfoLogger.Printf("Commande '%v' enregistrée !", v.Name)
	}

	logger.InfoLogger.Println("Enregistrement des commandes terminé !")
}
