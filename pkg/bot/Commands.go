package bot

import (
	"main/pkg/interactions/commandHandler"
	messageContextMenuHandler "main/pkg/interactions/contextMenuHandler/message"
	userContextMenuHandler "main/pkg/interactions/contextMenuHandler/user"
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

var (
	commands = []*discordgo.ApplicationCommand{
		// Slash commands
		{
			Name:        "ping",
			Description: "Savoir si le bot est en ligne",
		},
		{
			Name:        "birthday",
			Description: "Gérer les anniversaires",
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
			Name:         "config",
			Description:  "Configurer le bot",
			DMPermission: boolPtr(false),
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
			Name:         "talk",
			Description:  "Faire parler le bot",
			DMPermission: boolPtr(false),
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
			Name:         "quote",
			Description:  "Créer une citation",
			DMPermission: boolPtr(false),
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
		{
			Name:        "maintenance",
			Description: "[OWNER] Mettre le bot en mode maintenance",
		},
		{
			Name:        "motus",
			Description: "Lancer une partie de motus",
		},
		{
			Name:        "quiz",
			Description: "Gestionnaire de quiz",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "launch",
					Description: "Afficher une question de quiz aléatoire",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "create",
					Description: "Créer une question de quiz",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "question",
							Description: "Question à poser",
							Type:        discordgo.ApplicationCommandOptionString,
							MinLength:   intPtr(1),
							MaxLength:   256,
							Required:    true,
						},
						{
							Name:        "answer",
							Description: "Réponse à la question",
							Type:        discordgo.ApplicationCommandOptionString,
							MinLength:   intPtr(1),
							MaxLength:   256,
							Required:    true,
						},
						{
							Name:        "bad1",
							Description: "Mauvaise réponse 1",
							Type:        discordgo.ApplicationCommandOptionString,
							MinLength:   intPtr(1),
							MaxLength:   256,
							Required:    true,
						},
						{
							Name:        "bad2",
							Description: "Mauvaise réponse 2",
							Type:        discordgo.ApplicationCommandOptionString,
							MinLength:   intPtr(1),
							MaxLength:   256,
							Required:    true,
						},
						{
							Name:        "bad3",
							Description: "Mauvaise réponse 3",
							Type:        discordgo.ApplicationCommandOptionString,
							MinLength:   intPtr(1),
							MaxLength:   256,
							Required:    true,
						},
						{
							Name:        "category",
							Description: "Catégorie de la question",
							Type:        discordgo.ApplicationCommandOptionString,
							MinLength:   intPtr(1),
							MaxLength:   256,
							Required:    true,
						},
						{
							Name:        "difficulty",
							Description: "Difficulté de la question",
							Type:        discordgo.ApplicationCommandOptionString,
							MinLength:   intPtr(1),
							MaxLength:   256,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Facile",
									Value: "facile",
								},
								{
									Name:  "Normal",
									Value: "normal",
								},
								{
									Name:  "Difficile",
									Value: "difficile",
								},
							},
						},
					},
				},
				{
					Name:        "leaderboard",
					Description: "Afficher le classement des joueurs",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "type",
							Description: "Type de classement",
							Type:        discordgo.ApplicationCommandOptionString,
							MinLength:   intPtr(1),
							MaxLength:   256,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Ratio",
									Value: "ratio",
								},
								{
									Name:  "Bonnes réponses",
									Value: "goodAnswers",
								},
								{
									Name:  "Mauvaises réponses",
									Value: "badAnswers",
								},
							},
						},
					},
				},
				{
					Name:        "me",
					Description: "Afficher vos statistiques de quiz",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
		{
			Name:         "streamer",
			Description:  "Gérer les streamers",
			DMPermission: boolPtr(false),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "add",
					Description: "Ajouter un streamer",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "channel_name",
							Description: "Nom EXACT de la chaîne Twitch",
							Type:        discordgo.ApplicationCommandOptionString,
							MinLength:   intPtr(1),
							MaxLength:   256,
							Required:    true,
						},
						{
							Name:         "channel",
							Description:  "Canal où envoyer les notifications",
							Type:         discordgo.ApplicationCommandOptionChannel,
							Required:     true,
							ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText, discordgo.ChannelTypeGuildNews},
						},
						{
							Name:        "role",
							Description: "Rôle à mentionner lors de la notification",
							Type:        discordgo.ApplicationCommandOptionRole,
							Required:    false,
						},
					},
				},
				{
					Name:        "remove",
					Description: "Supprimer un streamer",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "channel_name",
							Description: "Nom EXACT de la chaîne Twitch",
							Type:        discordgo.ApplicationCommandOptionString,
							MinLength:   intPtr(1),
							MaxLength:   256,
							Required:    true,
						},
					},
				},
			},
		},

		// Context menu user commands
		{
			Name:         "Récupérer la photo de profil",
			Type:         discordgo.UserApplicationCommand,
			DMPermission: boolPtr(false),
		},
		{
			Name:         "Récupérer la bannière",
			Type:         discordgo.UserApplicationCommand,
			DMPermission: boolPtr(false),
		},

		// Context menu message commands
		{
			Name:         "Créer une citation",
			Type:         discordgo.MessageApplicationCommand,
			DMPermission: boolPtr(false),
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
		// Slash commands
		"ping":        commandHandler.PingHandler,
		"birthday":    commandHandler.BirthdayHandler,
		"blague":      commandHandler.BlagueHandler,
		"config":      commandHandler.ConfigHandler,
		"talk":        commandHandler.TalkHandler,
		"quote":       commandHandler.QuoteHandler,
		"debug":       commandHandler.DebugHandler,
		"maintenance": commandHandler.MaintenanceHandler,
		"motus":       commandHandler.MotusCommand,
		"quiz":        commandHandler.QuizHandler,
		"streamer":    commandHandler.StreamerHandler,

		// Context menu user commands
		"Récupérer la photo de profil": userContextMenuHandler.ProfilePictureContextMenuHandler,
		"Récupérer la bannière":        userContextMenuHandler.BannerContextMenuHandler,

		// Context menu message commands
		"Créer une citation": messageContextMenuHandler.CreateQuoteContextMenuHandler,
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
}

func GetCommandsForHelp() []*discordgo.ApplicationCommand {
	// Filter out commands that are not slash commands
	var slashCommands []*discordgo.ApplicationCommand
	for _, command := range commands {
		if command.Type == discordgo.ChatApplicationCommand {
			slashCommands = append(slashCommands, command)
		}
	}
	return slashCommands
}
