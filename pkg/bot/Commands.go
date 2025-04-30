package bot

import (
	"main/pkg/commandHandler"
	"main/pkg/contextMenuHandler"
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

func intPtr(i int) *int {
	return &i
}

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
		{
			Name:        "config",
			Description: "Configurer le bot",
			// Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "set",
					Description: "Configurer le bot",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "setting",
							Description: "Paramètre à configurer",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Canal d'anniversaire",
									Value: "birthdayChannel",
								},
								{
									Name:  "Canal des citations",
									Value: "quoteChannel",
								},
							},
						},
						{
							Name:         "channel",
							Description:  "Canal choisi",
							Type:         discordgo.ApplicationCommandOptionChannel,
							Required:     true,
							ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
						},
					},
				},
				{
					Name:        "get",
					Description: "Obtenir la configuration du bot",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "delete",
					Description: "Supprimer la configuration du bot",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "setting",
							Description: "Paramètre à supprimer",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Canal d'anniversaire",
									Value: "birthdayChannel",
								},
								{
									Name:  "Canal des citations",
									Value: "quoteChannel",
								},
							},
						},
					},
				},
			},
		},
		{
			Name:        "talk",
			Description: "Faire parler le bot",
			// Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "message",
					MinLength:   intPtr(1),
					MaxLength:   2000,
					Description: "Message à faire dire",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "mp",
					Description: "Dans le MPs de l'utilisateur ?",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    false,
				},
			},
		},
		{
			Name:        "quote",
			Description: "Créer une citation",
			// Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "citation",
					Description: "Citation à créer",
					Type:        discordgo.ApplicationCommandOptionString,
					MinLength:   intPtr(1),
					MaxLength:   256,
					Required:    true,
				},
				{
					Name:        "auteur",
					Description: "Auteur de la citation",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
				{
					Name:        "contexte",
					Description: "Contexte de la citation",
					Type:        discordgo.ApplicationCommandOptionString,
					MinLength:   intPtr(1),
					MaxLength:   256,
					Required:    false,
				},
			},
		},
		{
			Name:        "debug",
			Description: "[OWNER] Debug command",
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
		"config":   commandHandler.ConfigHandler,
		"talk":     commandHandler.TalkHandler,
		"quote":    commandHandler.QuoteHandler,
		"debug":    commandHandler.DebugHandler,

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

	// Register all commands at once using BulkOverwrite which is more efficient
	// This both removes old commands and adds the new ones in a single API call
	registeredCommands, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", commands)
	if err != nil {
		logger.ErrorLogger.Printf("Impossible d'enregistrer les commandes: %v", err)
		return
	}

	logger.InfoLogger.Printf("Total de %v commandes enregistrées avec succès!", len(registeredCommands))

	// Log each registered command for verification
	for _, cmd := range registeredCommands {
		logger.InfoLogger.Printf("Commande '%v' enregistrée !", cmd.Name)
	}
}
