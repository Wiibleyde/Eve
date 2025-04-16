package bot

import (
	"main/pkg/commandHandler"
	"main/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

var (
	Commands = []*discordgo.ApplicationCommand{
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
			Name:        "rolemanager",
			Description: "Gérer les rôles",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "create",
					Description: "Créer un gestionnaire de rôles",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "role",
							Description: "Rôle à ajouter au gestionnaire de rôles",
							Type:        discordgo.ApplicationCommandOptionRole,
							Required:    true,
						},
					},
				},
				{
					Name:        "delete",
					Description: "Supprimer un gestionnaire de rôles",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "id",
							Description: "ID du gestionnaire de rôles",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
						},
					},
				},
				{
					Name:        "addrole",
					Description: "Ajouter un rôle à un gestionnaire de rôles",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "id",
							Description: "ID du gestionnaire de rôles",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
						},
						{
							Name:        "role",
							Description: "Rôle à ajouter au gestionnaire de rôles",
							Type:        discordgo.ApplicationCommandOptionRole,
							Required:    true,
						},
					},
				},
				{
					Name:        "removerole",
					Description: "Supprimer un rôle d'un gestionnaire de rôles",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "id",
							Description: "ID du gestionnaire de rôles",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
						},
						{
							Name:        "role",
							Description: "Rôle à supprimer du gestionnaire de rôles",
							Type:        discordgo.ApplicationCommandOptionRole,
							Required:    true,
						},
					},
				},
			},
		},
	}
	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
		"ping": commandHandler.PingHandler,
		//TODO: Add role manager handler
		"birthday": commandHandler.BirthdayHandler,
	}
)

// CheckCommandHandlers check if every command has a handler
func CheckCommandHandlers() {
	for i, v := range Commands {
		if _, ok := CommandHandlers[v.Name]; !ok {
			Commands = append(Commands[:i], Commands[i+1:]...)
			logger.WarningLogger.Println("Removed command", v.Name)
			continue
		}
	}
}

func RegisterCommands(s *discordgo.Session) {
	CheckCommandHandlers()

	// Remove previous commands
	_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", Commands)
	if err != nil {
		logger.ErrorLogger.Printf("Impossible de supprimer les commandes précédentes: %v", err)
		return
	}
	logger.InfoLogger.Println("Suppression des commandes précédentes terminée !")

	// Register new commands
	registeredCommands := make([]*discordgo.ApplicationCommand, len(Commands))

	for i, v := range Commands {
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
