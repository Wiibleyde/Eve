package loto

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/bot/router"
	"Eve/internal/database/ent"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func HandleBuyButton(e *events.ComponentInteractionCreate, args []string) {
	game, ok := requireActiveGame(e, args)
	if !ok {
		return
	}

	placeholder := "Ex: 3"
	if game.MaxTicketsPerPurchase > 0 {
		placeholder = fmt.Sprintf("Ex: 3 (max %d)", game.MaxTicketsPerPurchase)
	}

	modal := discord.NewModalCreate(router.BuildCustomID(modalBuy, game.ID), "Ajouter des tickets").
		AddLabel("Nom du joueur (sensible à la casse)",
			discord.NewShortTextInput(inputPlayerName).
				WithRequired(true).
				WithMaxLength(50).
				WithPlaceholder("Ex: John Doe")).
		AddLabel("Nombre de tickets",
			discord.NewShortTextInput(inputTicketCount).
				WithRequired(true).
				WithMaxLength(5).
				WithPlaceholder(placeholder))

	if err := e.Modal(modal); err != nil {
		logger.Error("loto: opening buy modal", "error", err)
	}
}

func HandleRemoveButton(e *events.ComponentInteractionCreate, args []string) {
	game, ok := requireActiveGame(e, args)
	if !ok {
		return
	}

	modal := discord.NewModalCreate(router.BuildCustomID(modalRemove, game.ID), "Retirer des tickets").
		AddLabel("Nom du joueur (sensible à la casse)",
			discord.NewShortTextInput(inputPlayerName).
				WithRequired(true).
				WithMaxLength(50).
				WithPlaceholder("Ex: John Doe")).
		AddLabel("Nombre de tickets à retirer",
			discord.NewShortTextInput(inputTicketCount).
				WithRequired(true).
				WithMaxLength(5).
				WithPlaceholder("Ex: 1"))

	if err := e.Modal(modal); err != nil {
		logger.Error("loto: opening remove modal", "error", err)
	}
}

func HandleEditPlayerButton(e *events.ComponentInteractionCreate, args []string) {
	game, ok := requireActiveGame(e, args)
	if !ok {
		return
	}
	if !requireAdmin(e, e.Member()) {
		return
	}

	modal := discord.NewModalCreate(router.BuildCustomID(modalEditPlayer, game.ID), "Corriger un nom").
		AddLabel("Nom actuel",
			discord.NewShortTextInput(inputOldName).
				WithRequired(true).
				WithMaxLength(50)).
		AddLabel("Nouveau nom",
			discord.NewShortTextInput(inputNewName).
				WithRequired(true).
				WithMaxLength(50))

	if err := e.Modal(modal); err != nil {
		logger.Error("loto: opening edit player modal", "error", err)
	}
}

func HandleDrawButton(e *events.ComponentInteractionCreate, args []string) {
	game, ok := requireActiveGame(e, args)
	if !ok {
		return
	}
	if !requireAdmin(e, e.Member()) {
		return
	}

	snap, err := loadSnapshot(context.Background(), game.ID)
	if err != nil {
		logger.Error("loto: loading game for draw", "game", game.ID, "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de la lecture du loto."))
		return
	}
	if snap == nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Ce loto n'existe plus."))
		return
	}
	if len(snap.tickets) == 0 {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Aucun ticket vendu, impossible de tirer au sort."))
		return
	}
	if len(snap.prizes) == 0 {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Aucun gain configuré pour ce loto."))
		return
	}
	if !requireEnoughPlayers(e, snap) {
		return
	}

	embed := embeds.BaseEmbed()
	embed.Color = 0xFFA500
	embed.Title = "Confirmer le tirage"
	embed.Description = fmt.Sprintf(
		"Vous êtes sur le point de tirer au sort **%d gain(s)** parmi **%d ticket(s)** pour le loto **%s**.\n\n"+
			"⚠️ Cette action est **définitive** : le loto sera clôturé et plus aucun ticket ne pourra être vendu.",
		len(snap.prizes), len(snap.tickets), snap.game.Name,
	)

	err = e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Components: []discord.LayoutComponent{
			discord.NewActionRow(
				discord.NewDangerButton("Confirmer le tirage", router.BuildCustomID(buttonDrawConfirm, game.ID)).
					WithEmoji(discord.NewComponentEmoji("🎲")),
			),
		},
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		logger.Error("loto: sending draw confirmation", "error", err)
	}
}

func HandleDrawConfirmButton(e *events.ComponentInteractionCreate, args []string) {
	game, ok := requireActiveGame(e, args)
	if !ok {
		return
	}
	if !requireAdmin(e, e.Member()) {
		return
	}

	ctx := context.Background()
	snap, err := loadSnapshot(ctx, game.ID)
	if err != nil {
		logger.Error("loto: loading game for draw", "game", game.ID, "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de la lecture du loto."))
		return
	}
	if snap == nil {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Ce loto n'existe plus."))
		return
	}
	if len(snap.tickets) == 0 {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Aucun ticket vendu, impossible de tirer au sort."))
		return
	}
	if len(snap.prizes) == 0 {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Aucun gain configuré pour ce loto."))
		return
	}
	if !requireEnoughPlayers(e, snap) {
		return
	}

	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("loto: deferring draw", "error", err)
		return
	}

	assignments := drawAssignments(snap)
	if err := applyDraw(ctx, game.ID, assignments); err != nil {
		if errors.Is(err, errDrawAlreadyDone) {
			followupError(e, "Ce loto est terminé.")
			return
		}
		logger.Error("loto: applying draw", "game", game.ID, "error", err)
		followupError(e, "Erreur lors de l'enregistrement du tirage.")
		return
	}

	final, err := loadSnapshot(ctx, game.ID)
	if err != nil || final == nil {
		logger.Error("loto: reloading game after draw", "game", game.ID, "error", err)
		followupError(e, "Le tirage a été enregistré mais l'affichage n'a pas pu être mis à jour.")
		return
	}

	editPublicMessage(e.Client(), final)
	announceDraw(e.Client(), final)

	summary := embeds.SuccessEmbed(fmt.Sprintf(
		"🎉 Tirage effectué !\n%s\n\nCagnotte totale: **%d$**.",
		strings.Join(winnerLines(final), "\n"), final.pot(),
	))
	updateResponse(e, summary)
}

func (s *snapshot) distinctTicketHolders() int {
	holders := make(map[string]struct{}, len(s.players))
	for _, t := range s.tickets {
		holders[t.PlayerID] = struct{}{}
	}
	return len(holders)
}

func requireEnoughPlayers(r helpers.EphemeralResponder, snap *snapshot) bool {
	holders := snap.distinctTicketHolders()
	if len(snap.prizes) <= holders {
		return true
	}
	helpers.RespondEphemeralEmbed(r, embeds.ErrorEmbed(fmt.Sprintf(
		"Impossible de tirer au sort : %d gain(s) pour seulement %d joueur(s) ayant des tickets. "+
			"Chaque joueur ne peut gagner qu'un seul gain.",
		len(snap.prizes), holders)))
	return false
}

func drawAssignments(snap *snapshot) []assignment {
	names := snap.playerNames()
	remaining := make([]*ent.LotoTicket, len(snap.tickets))
	copy(remaining, snap.tickets)

	assignments := make([]assignment, 0, len(snap.prizes))
	for _, prize := range snap.prizes {
		if len(remaining) == 0 {
			break
		}
		i := rand.IntN(len(remaining))
		ticket := remaining[i]
		withoutWinner := remaining[:0]
		for _, t := range remaining {
			if t.PlayerID != ticket.PlayerID {
				withoutWinner = append(withoutWinner, t)
			}
		}
		remaining = withoutWinner

		name := names[ticket.PlayerID]
		if name == "" {
			name = "Nom inconnu"
		}
		assignments = append(assignments, assignment{
			prize:      prize,
			playerID:   ticket.PlayerID,
			playerName: name,
			ticket:     ticket,
		})
	}
	return assignments
}

func winnerLines(snap *snapshot) []string {
	names := snap.playerNames()
	lines := make([]string, 0, len(snap.prizes))
	for i, prize := range snap.prizes {
		rank := i + 1
		if prize.WinnerPlayerID == "" {
			lines = append(lines, fmt.Sprintf("#%d %s → Non attribué", rank, prize.Label))
			continue
		}
		winner := names[prize.WinnerPlayerID]
		if winner == "" {
			winner = "Nom inconnu"
		}
		ticketInfo := ""
		if prize.WinningTicketNumber > 0 {
			ticketInfo = fmt.Sprintf(" (ticket n°%d)", prize.WinningTicketNumber)
		}
		lines = append(lines, fmt.Sprintf("#%d %s → **%s**%s", rank, prize.Label, winner, ticketInfo))
	}
	return lines
}

func announceDraw(client *bot.Client, snap *snapshot) {
	channelID, ok := parseSnowflake(snap.game.ChannelID, "channel")
	if !ok {
		return
	}

	embed := embeds.BaseEmbed()
	embed.Color = 0xFFD700
	embed.Title = fmt.Sprintf("🎉 Résultats du loto: %s", snap.game.Name)
	embed.Description = truncate(fmt.Sprintf(
		"%s\n\nTickets vendus: **%d**\nCagnotte totale: **%d$**",
		strings.Join(winnerLines(snap), "\n"), len(snap.tickets), snap.pot(),
	), maxEmbedDescription)

	if _, err := client.Rest.CreateMessage(channelID, discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	}); err != nil {
		logger.Error("loto: announcing draw", "game", snap.game.ID, "error", err)
	}
}

func requireActiveGame(r helpers.EphemeralResponder, args []string) (*ent.LotoGame, bool) {
	if len(args) == 0 || args[0] == "" {
		helpers.RespondEphemeralEmbed(r, embeds.ErrorEmbed("Loto introuvable."))
		return nil, false
	}

	game, err := gameByID(context.Background(), args[0])
	if err != nil {
		logger.Error("loto: loading game", "game", args[0], "error", err)
		helpers.RespondEphemeralEmbed(r, embeds.ErrorEmbed("Erreur lors de la lecture du loto."))
		return nil, false
	}
	if game == nil {
		helpers.RespondEphemeralEmbed(r, embeds.ErrorEmbed("Ce loto n'existe plus."))
		return nil, false
	}
	if !game.Active {
		helpers.RespondEphemeralEmbed(r, embeds.ErrorEmbed("Ce loto est terminé."))
		return nil, false
	}
	return game, true
}

func isAdmin(member *discord.ResolvedMember) bool {
	return member != nil && member.Permissions.Has(discord.PermissionManageMessages)
}

func requireAdmin(r helpers.EphemeralResponder, member *discord.ResolvedMember) bool {
	if isAdmin(member) {
		return true
	}
	helpers.RespondEphemeralEmbed(r, embeds.ErrorEmbed(msgNoPermission))
	return false
}

func refreshPublicMessage(client *bot.Client, gameID string) {
	snap, err := loadSnapshot(context.Background(), gameID)
	if err != nil {
		logger.Error("loto: reloading game for refresh", "game", gameID, "error", err)
		return
	}
	if snap == nil {
		return
	}
	editPublicMessage(client, snap)
}

func editPublicMessage(client *bot.Client, snap *snapshot) {
	game := snap.game
	if game.ChannelID == "" || game.MessageID == "" {
		logger.Debug("loto: no public message stored for game", "game", game.ID)
		return
	}
	channelID, ok := parseSnowflake(game.ChannelID, "channel")
	if !ok {
		return
	}
	messageID, ok := parseSnowflake(game.MessageID, "message")
	if !ok {
		return
	}

	embed := buildEmbed(snap)
	components := buildComponents(game.ID)
	if !game.Active {
		embed.Fields = append(embed.Fields, discord.EmbedField{
			Name:  "📊 Ventes par vendeur",
			Value: sellerSummary(snap.tickets, game.TicketPrice),
		})
		components = []discord.LayoutComponent{}
	}

	embedList := []discord.Embed{embed}
	if _, err := client.Rest.UpdateMessage(channelID, messageID, discord.MessageUpdate{
		Embeds:     &embedList,
		Components: &components,
	}); err != nil {
		logger.Error("loto: updating public message", "game", game.ID, "error", err)
	}
}

func parseSnowflake(raw string, kind string) (snowflake.ID, bool) {
	if raw == "" {
		return 0, false
	}
	id, err := snowflake.Parse(raw)
	if err != nil {
		logger.Error("loto: parsing id", "kind", kind, "value", raw, "error", err)
		return 0, false
	}
	return id, true
}

func updateResponse(e *events.ComponentInteractionCreate, embed discord.Embed) {
	embedList := []discord.Embed{embed}
	if _, err := e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.MessageUpdate{
		Embeds: &embedList,
	}); err != nil {
		logger.Error("loto: updating interaction response", "error", err)
	}
}

func followupError(e *events.ComponentInteractionCreate, message string) {
	updateResponse(e, embeds.ErrorEmbed(message))
}
