package loto

import (
	"context"
	"fmt"
	"strings"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/ui"
	"Eve/internal/database/ent"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const (
	maxGameNameLength   = 50
	maxPlayerNameLength = 50
	maxPrizeLabelLen    = 255
	maxPrizeCount       = 10
	defaultTicketPrice  = 500
	maxCooldownMinutes  = 10080
	maxTicketsPerBuy    = 1000
)

const msgNoPermission = "Vous n'avez pas la permission de gérer les messages, action refusée."

const msgTooLong = "%s est trop long (%d caractères maximum, les accents et emojis comptant pour plusieurs)."

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "loto",
		Description: "Gestion du loto RP",
		Contexts:    []discord.InteractionContextType{discord.InteractionContextTypeGuild},
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "create",
				Description: "[ADMIN] Créer un nouveau loto",
				Options:     createOptions(),
			},
		},
	},
}

func createOptions() []discord.ApplicationCommandOption {
	opts := []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "name",
			Description: "Nom du loto",
			Required:    true,
			MaxLength:   intPtr(maxGameNameLength),
		},
		discord.ApplicationCommandOptionString{
			Name:        "prize1",
			Description: "Gain n°1 (obligatoire)",
			Required:    true,
			MaxLength:   intPtr(maxPrizeLabelLen),
		},
		discord.ApplicationCommandOptionInt{
			Name:        "ticketprice",
			Description: "Prix d'un ticket (défaut: 500)",
			MinValue:    intPtr(0),
		},
		discord.ApplicationCommandOptionInt{
			Name:        "cooldown",
			Description: "Minutes entre deux achats pour un même joueur (défaut: 0)",
			MinValue:    intPtr(0),
			MaxValue:    intPtr(maxCooldownMinutes),
		},
		discord.ApplicationCommandOptionInt{
			Name:        "maxtickets",
			Description: "Nombre max de tickets par achat (obligatoire si un cooldown est défini)",
			MinValue:    intPtr(1),
			MaxValue:    intPtr(maxTicketsPerBuy),
		},
	}

	for i := 2; i <= maxPrizeCount; i++ {
		opts = append(opts, discord.ApplicationCommandOptionString{
			Name:        fmt.Sprintf("prize%d", i),
			Description: fmt.Sprintf("Gain n°%d (optionnel)", i),
			MaxLength:   intPtr(maxPrizeLabelLen),
		})
	}
	return opts
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()
	if data.SubCommandName == nil {
		return
	}
	switch *data.SubCommandName {
	case "create":
		handleCreate(e)
	default:
		helpers.RespondEphemeralCard(e, ui.Error("Sous-commande inconnue."))
	}
}

func handleCreate(e *events.ApplicationCommandInteractionCreate) {
	if e.GuildID() == nil {
		helpers.RespondEphemeralCard(e, ui.Error("Cette commande doit être utilisée dans un serveur."))
		return
	}
	if !helpers.RequirePermission(e, discord.PermissionManageMessages, msgNoPermission) {
		return
	}

	data := e.SlashCommandInteractionData()

	name := strings.TrimSpace(data.String("name"))
	if name == "" {
		helpers.RespondEphemeralCard(e, ui.Error("Le nom du loto ne peut pas être vide."))
		return
	}
	if len(name) > maxGameNameLength {
		helpers.RespondEphemeralCard(e, ui.Error(fmt.Sprintf(msgTooLong, "Le nom du loto", maxGameNameLength)))
		return
	}

	ticketPrice := defaultTicketPrice
	if v, ok := data.OptInt("ticketprice"); ok {
		ticketPrice = v
	}
	if ticketPrice < 0 {
		helpers.RespondEphemeralCard(e, ui.Error("Le prix du ticket doit être un entier positif ou nul."))
		return
	}

	cooldown, _ := data.OptInt("cooldown")
	if cooldown < 0 {
		helpers.RespondEphemeralCard(e, ui.Error("Le cooldown doit être positif."))
		return
	}

	maxTickets, hasMaxTickets := data.OptInt("maxtickets")
	if hasMaxTickets && maxTickets <= 0 {
		helpers.RespondEphemeralCard(e, ui.Error("Le nombre maximum de tickets par achat doit être strictement positif."))
		return
	}
	if cooldown > 0 && (!hasMaxTickets || maxTickets <= 0) {
		helpers.RespondEphemeralCard(e, ui.Error("Vous devez définir `maxtickets` (strictement positif) lorsque vous définissez un cooldown."))
		return
	}
	if !hasMaxTickets {
		maxTickets = 0
	}

	prizes := make([]string, 0, maxPrizeCount)
	for i := 1; i <= maxPrizeCount; i++ {
		raw, ok := data.OptString(fmt.Sprintf("prize%d", i))
		if !ok {
			continue
		}
		label := strings.TrimSpace(raw)
		if label == "" {
			continue
		}
		if len(label) > maxPrizeLabelLen {
			helpers.RespondEphemeralCard(e, ui.Error(fmt.Sprintf(msgTooLong, fmt.Sprintf("Le gain n°%d", i), maxPrizeLabelLen)))
			return
		}
		prizes = append(prizes, label)
	}
	if len(prizes) == 0 {
		helpers.RespondEphemeralCard(e, ui.Error("Vous devez fournir au moins un gain."))
		return
	}

	ctx := context.Background()
	guildID := e.GuildID().String()

	existing, err := activeGame(ctx, guildID)
	if err != nil {
		logger.Error("loto: checking active game", "error", err)
		helpers.RespondEphemeralCard(e, ui.Error("Erreur lors de la vérification des lotos en cours."))
		return
	}
	if existing != nil {
		helpers.RespondEphemeralCard(e, ui.Error(fmt.Sprintf("Un loto est déjà en cours sur ce serveur (**%s**). Terminez-le avant d'en créer un nouveau.", existing.Name)))
		return
	}

	snap, err := createGame(ctx, guildID, name, ticketPrice, cooldown, maxTickets, prizes)
	if err != nil {
		if ent.IsConstraintError(err) {
			helpers.RespondEphemeralCard(e, ui.Error(
				"Un loto est déjà en cours sur ce serveur. Terminez-le avant d'en créer un nouveau."))
			return
		}
		logger.Error("loto: creating game", "error", err)
		helpers.RespondEphemeralCard(e, ui.Error("Erreur lors de la création du loto."))
		return
	}

	card := buildCard(snap).Subtext(fmt.Sprintf("Nouveau loto créé par <@%s>", e.User().ID))
	if err := e.CreateMessage(card.MessageCreate()); err != nil {
		logger.Error("loto: posting public message", "error", err)
		if delErr := deleteGame(ctx, snap.game.ID); delErr != nil {
			logger.Error("loto: rolling back game creation", "game", snap.game.ID, "error", delErr)
		}
		helpers.RespondEphemeralCard(e, ui.Error(
			"Impossible de publier le message du loto dans ce salon (permissions manquantes ?). Le loto n'a pas été créé."))
		return
	}

	msg, err := e.Client().Rest.GetInteractionResponse(e.ApplicationID(), e.Token())
	if err != nil {
		logger.Warn("loto: could not resolve public message", "game", snap.game.ID, "error", err)
		warnMessageUnlinked(e)
		return
	}
	if err := setGameMessage(ctx, snap.game.ID, msg.ChannelID.String(), msg.ID.String()); err != nil {
		logger.Warn("loto: storing public message reference", "game", snap.game.ID, "error", err)
		warnMessageUnlinked(e)
	}
}

func warnMessageUnlinked(e *events.ApplicationCommandInteractionCreate) {
	helpers.RespondFollowupEphemeral(e.Client(), e.ApplicationID(), e.Token(),
		"⚠️ Le loto a bien été créé, mais son message public n'a pas pu être mémorisé : "+
			"son affichage ne se mettra plus à jour et les résultats du tirage ne seront pas annoncés dans le salon.")
}

func intPtr(v int) *int { return &v }
