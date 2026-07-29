package loto

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"Eve/internal/bot/embeds"
	"Eve/internal/bot/helpers"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/events"
)

func HandleBuyModal(e *events.ModalSubmitInteractionCreate, args []string) {
	game, ok := requireActiveGame(e, args)
	if !ok {
		return
	}

	name := strings.TrimSpace(e.Data.Text(inputPlayerName))
	if name == "" {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Le nom du joueur ne peut pas être vide."))
		return
	}
	if len(name) > maxPlayerNameLength {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(msgTooLong, "Le nom du joueur", maxPlayerNameLength)))
		return
	}

	count, ok := parseCount(e, e.Data.Text(inputTicketCount))
	if !ok {
		return
	}
	if game.MaxTicketsPerPurchase > 0 && count > game.MaxTicketsPerPurchase {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(
			"Un joueur ne peut pas acheter plus de **%d** ticket(s) par achat sur ce loto.", game.MaxTicketsPerPurchase)))
		return
	}

	ctx := context.Background()
	player, err := findPlayer(ctx, game.ID, name)
	if err != nil {
		logger.Error("loto: looking up player", "game", game.ID, "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de la lecture du joueur."))
		return
	}

	if player != nil && game.CooldownMinutes > 0 {
		nextAllowed := player.LastPlay.Add(time.Duration(game.CooldownMinutes) * time.Minute)
		if remaining := time.Until(nextAllowed); remaining > 0 {
			helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(
				"**%s** doit patienter encore %s avant de pouvoir racheter des tickets.", name, formatRemaining(remaining))))
			return
		}
	}

	sellerID := e.User().ID.String()
	if err := addTickets(ctx, game.ID, player, name, sellerID, count); err != nil {
		logger.Error("loto: adding tickets", "game", game.ID, "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors de l'ajout des tickets."))
		return
	}

	helpers.RespondEphemeralEmbed(e, embeds.SuccessEmbed(fmt.Sprintf(
		"%d ticket(s) ajouté(s) pour **%s** par <@%s>.\nMontant total: **%d$**",
		count, name, sellerID, count*game.TicketPrice)))

	refreshPublicMessage(e.Client(), game.ID)
}

func HandleRemoveModal(e *events.ModalSubmitInteractionCreate, args []string) {
	game, ok := requireActiveGame(e, args)
	if !ok {
		return
	}

	name := strings.TrimSpace(e.Data.Text(inputPlayerName))
	if name == "" {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Le nom du joueur ne peut pas être vide."))
		return
	}

	count, ok := parseCount(e, e.Data.Text(inputTicketCount))
	if !ok {
		return
	}

	sellerFilter := ""
	if !isAdmin(e.Member()) {
		sellerFilter = e.User().ID.String()
	}

	removed, err := removeTickets(context.Background(), game.ID, name, count, sellerFilter)
	switch {
	case errors.Is(err, errPlayerNotFound):
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf("Le joueur **%s** n'existe pas dans ce loto.", name)))
		return
	case errors.Is(err, errNoTickets):
		if sellerFilter != "" {
			helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(
				"Vous n'avez vendu aucun ticket à **%s**. Seul un administrateur peut retirer les tickets vendus par quelqu'un d'autre.", name)))
			return
		}
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf("**%s** ne possède aucun ticket.", name)))
		return
	case err != nil:
		logger.Error("loto: removing tickets", "game", game.ID, "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors du retrait des tickets."))
		return
	}

	helpers.RespondEphemeralEmbed(e, embeds.SuccessEmbed(fmt.Sprintf(
		"%d ticket(s) retiré(s) pour **%s**.", removed, name)))

	refreshPublicMessage(e.Client(), game.ID)
}

func HandleEditPlayerModal(e *events.ModalSubmitInteractionCreate, args []string) {
	game, ok := requireActiveGame(e, args)
	if !ok {
		return
	}
	if !requireAdmin(e, e.Member()) {
		return
	}

	oldName := strings.TrimSpace(e.Data.Text(inputOldName))
	newName := strings.TrimSpace(e.Data.Text(inputNewName))

	if oldName == "" || newName == "" {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Les noms ne peuvent pas être vides."))
		return
	}
	if oldName == newName {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Le nouveau nom doit être différent de l'ancien."))
		return
	}
	if len(newName) > maxPlayerNameLength {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(msgTooLong, "Le nouveau nom", maxPlayerNameLength)))
		return
	}

	err := renamePlayer(context.Background(), game.ID, oldName, newName)
	switch {
	case errors.Is(err, errPlayerNotFound):
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf("Le joueur **%s** n'existe pas dans ce loto.", oldName)))
		return
	case errors.Is(err, errNameTaken):
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf("Un joueur nommé **%s** existe déjà dans ce loto.", newName)))
		return
	case err != nil:
		logger.Error("loto: renaming player", "game", game.ID, "error", err)
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Erreur lors du renommage."))
		return
	}

	helpers.RespondEphemeralEmbed(e, embeds.SuccessEmbed(fmt.Sprintf("**%s** a été renommé en **%s**.", oldName, newName)))

	refreshPublicMessage(e.Client(), game.ID)
}

func parseCount(e *events.ModalSubmitInteractionCreate, raw string) (int, bool) {
	count, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || count <= 0 {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed("Le nombre de tickets doit être un entier strictement positif."))
		return 0, false
	}
	if count > maxTicketsPerBuy {
		helpers.RespondEphemeralEmbed(e, embeds.ErrorEmbed(fmt.Sprintf(
			"Le nombre de tickets ne peut pas dépasser %d en une seule opération.", maxTicketsPerBuy)))
		return 0, false
	}
	return count, true
}

func formatRemaining(d time.Duration) string {
	total := int(math.Ceil(d.Minutes()))
	if total <= 0 {
		return "moins d'une minute"
	}

	hours := total / 60
	minutes := total % 60

	parts := make([]string, 0, 2)
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", hours, pluralize(hours, "heure", "heures")))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", minutes, pluralize(minutes, "minute", "minutes")))
	}
	if len(parts) == 0 {
		return "moins d'une minute"
	}
	return strings.Join(parts, " et ")
}
