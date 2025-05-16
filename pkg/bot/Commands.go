package bot

import (
	"main/pkg/interactions/commandHandler"
	"main/pkg/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

var (
	commands = []discord.ApplicationCommandCreate{
		// Slash commands
		discord.SlashCommandCreate{
			Name:        "ping",
			Description: "Savoir si le bot est en ligne",
		},
		discord.SlashCommandCreate{
			Name:        "birthday",
			Description: "Gérer les anniversaires",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "set",
					Description: "Définir votre date de naissance",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "date",
							Description: "Date de naissance au format YYYY-MM-DD",
							Required:    true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "get",
					Description: "Obtenir votre date de naissance",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionUser{
							Name:        "user",
							Description: "Utilisateur dont vous voulez connaître la date de naissance",
							Required:    false,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "delete",
					Description: "Supprimer votre date de naissance",
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "list",
					Description: "Lister les anniversaires (dans le serveur)",
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "blague",
			Description: "Obtenir une blague",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "type",
					Description: "Type de blague (certains types ont été censurés)",
					Required:    true,
					Choices: []discord.ApplicationCommandOptionChoiceString{
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
		discord.SlashCommandCreate{
			Name:         "config",
			Description:  "Configurer le bot",
			DMPermission: boolPtr(false),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "set",
					Description: "Configurer le bot",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "setting",
							Description: "Paramètre à configurer",
							Required:    true,
							Choices: []discord.ApplicationCommandOptionChoiceString{
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
						discord.ApplicationCommandOptionChannel{
							Name:         "channel",
							Description:  "Canal choisi",
							Required:     true,
							ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "get",
					Description: "Obtenir la configuration du bot",
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "delete",
					Description: "Supprimer la configuration du bot",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "setting",
							Description: "Paramètre à supprimer",
							Required:    true,
							Choices: []discord.ApplicationCommandOptionChoiceString{
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
		discord.SlashCommandCreate{
			Name:         "talk",
			Description:  "Faire parler le bot",
			DMPermission: boolPtr(false),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "message",
					Description: "Message à faire dire",
					Required:    true,
					MinLength:   intPtr(1),
					MaxLength:   intPtr(2000),
				},
				discord.ApplicationCommandOptionUser{
					Name:        "mp",
					Description: "Dans le MPs de l'utilisateur ?",
					Required:    false,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:         "quote",
			Description:  "Créer une citation",
			DMPermission: boolPtr(false),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "citation",
					Description: "Citation à créer",
					Required:    true,
					MinLength:   intPtr(1),
					MaxLength:   intPtr(256),
				},
				discord.ApplicationCommandOptionUser{
					Name:        "auteur",
					Description: "Auteur de la citation",
					Required:    true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "contexte",
					Description: "Contexte de la citation",
					Required:    false,
					MinLength:   intPtr(1),
					MaxLength:   intPtr(256),
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "debug",
			Description: "[OWNER] Debug command",
		},
		discord.SlashCommandCreate{
			Name:        "maintenance",
			Description: "[OWNER] Mettre le bot en mode maintenance",
		},
		discord.SlashCommandCreate{
			Name:        "motus",
			Description: "Lancer une partie de motus",
		},
		discord.SlashCommandCreate{
			Name:        "quiz",
			Description: "Gestionnaire de quiz",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "launch",
					Description: "Afficher une question de quiz aléatoire",
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "create",
					Description: "Créer une question de quiz",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "question",
							Description: "Question à poser",
							MinLength:   intPtr(1),
							MaxLength:   intPtr(256),
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "answer",
							Description: "Réponse à la question",
							MinLength:   intPtr(1),
							MaxLength:   intPtr(256),
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "bad1",
							Description: "Mauvaise réponse 1",
							MinLength:   intPtr(1),
							MaxLength:   intPtr(256),
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "bad2",
							Description: "Mauvaise réponse 2",
							MinLength:   intPtr(1),
							MaxLength:   intPtr(256),
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "bad3",
							Description: "Mauvaise réponse 3",
							MinLength:   intPtr(1),
							MaxLength:   intPtr(256),
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "category",
							Description: "Catégorie de la question",
							MinLength:   intPtr(1),
							MaxLength:   intPtr(256),
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "difficulty",
							Description: "Difficulté de la question",
							MinLength:   intPtr(1),
							MaxLength:   intPtr(256),
							Required:    true,
							Choices: []discord.ApplicationCommandOptionChoiceString{
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
				discord.ApplicationCommandOptionSubCommand{
					Name:        "leaderboard",
					Description: "Afficher le classement des joueurs",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "type",
							Description: "Type de classement",
							Required:    true,
							MinLength:   intPtr(1),
							MaxLength:   intPtr(256),
							Choices: []discord.ApplicationCommandOptionChoiceString{
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
				discord.ApplicationCommandOptionSubCommand{
					Name:        "me",
					Description: "Afficher vos statistiques de quiz",
				},
			},
		},
		discord.SlashCommandCreate{
			Name:         "streamer",
			Description:  "Gérer les streamers",
			DMPermission: boolPtr(false),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "add",
					Description: "Ajouter un streamer",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "channel_name",
							Description: "Nom EXACT de la chaîne Twitch",
							Required:    true,
							MinLength:   intPtr(1),
							MaxLength:   intPtr(256),
						},
						discord.ApplicationCommandOptionChannel{
							Name:         "channel",
							Description:  "Canal où envoyer les notifications",
							Required:     true,
							ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText, discord.ChannelTypeGuildNews},
						},
						discord.ApplicationCommandOptionRole{
							Name:        "role",
							Description: "Rôle à mentionner lors de la notification",
							Required:    false,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "remove",
					Description: "Supprimer un streamer",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "channel_name",
							Description: "Nom EXACT de la chaîne Twitch",
							Required:    true,
							MinLength:   intPtr(1),
							MaxLength:   intPtr(256),
						},
					},
				},
			},
		},
		discord.UserCommandCreate{
			Name:         "Récupérer la photo de profil",
			DMPermission: boolPtr(false),
		},
		discord.UserCommandCreate{
			Name:         "Récupérer la bannière",
			DMPermission: boolPtr(false),
		},
		discord.MessageCommandCreate{
			Name:         "Créer une citation",
			DMPermission: boolPtr(false),
		},
	}
	commandHandlers = map[string]func(event *events.ApplicationCommandInteractionCreate){
		// Slash commands
		"ping": commandHandler.PingHandler,
		// "birthday":    commandHandler.BirthdayHandler,
		// "blague":      commandHandler.BlagueHandler,
		// "config":      commandHandler.ConfigHandler,
		// "talk":        commandHandler.TalkHandler,
		// "quote":       commandHandler.QuoteHandler,
		// "debug":       commandHandler.DebugHandler,
		// "maintenance": commandHandler.MaintenanceHandler,
		// "motus":       commandHandler.MotusCommand,
		// "quiz":        commandHandler.QuizHandler,
		// "streamer":    commandHandler.StreamerHandler,

		// // Context menu user commands
		// "Récupérer la photo de profil": userContextMenuHandler.ProfilePictureContextMenuHandler,
		// "Récupérer la bannière":        userContextMenuHandler.BannerContextMenuHandler,

		// // Context menu message commands
		// "Créer une citation": messageContextMenuHandler.CreateQuoteContextMenuHandler,
	}
)

// checkCommandHandlers check if every command has a handler
func checkCommandHandlers() {
	for i, v := range commands {
		var name string
		switch cmd := v.(type) {
		case discord.SlashCommandCreate:
			name = cmd.Name
		case discord.UserCommandCreate:
			name = cmd.Name
		case discord.MessageCommandCreate:
			name = cmd.Name
		}

		if _, ok := commandHandlers[name]; !ok {
			commands = append(commands[:i], commands[i+1:]...)
			logger.WarningLogger.Println("Removed command", name)
			continue
		}
	}
}

func registerCommands(client bot.Client) {
	checkCommandHandlers()

	var err error
	if _, err = client.Rest().SetGlobalCommands(client.ApplicationID(), commands); err != nil {
		logger.ErrorLogger.Panicln("[PANIC] Impossible d'enregistrer les commandes globales,", err)
		return
	}

	logger.InfoLogger.Printf("Total de %v commandes enregistrées avec succès!", len(commands))
}
