package loto

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"Eve/internal/bot/helpers"
	"Eve/internal/bot/router"
	"Eve/internal/bot/ui"
	"Eve/internal/database/ent"
	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const (
	maxSelectLabelLength = 100
	maxSelectOptions     = 25
)

func HandlePrizesButton(e *events.ComponentInteractionCreate, args []string) {
	game, ok := requireActiveGame(e, args)
	if !ok {
		return
	}
	if !requireAdmin(e, e.Member()) {
		return
	}

	snap, ok := loadPrizeSnapshot(e, game.ID)
	if !ok {
		return
	}
	helpers.RespondEphemeralCard(e, prizeMenuCard(snap, "", ""))
}

func HandlePrizeSelect(e *events.ComponentInteractionCreate, args []string) {
	game, ok := requireActiveGame(e, args)
	if !ok {
		return
	}
	if !requireAdmin(e, e.Member()) {
		return
	}

	values := e.StringSelectMenuInteractionData().Values
	if len(values) == 0 {
		return
	}

	if values[0] == prizeAddValue {
		openPrizeModal(e, router.BuildCustomID(modalPrizeAdd, game.ID), "Ajouter un lot", "", "")
		return
	}

	snap, ok := loadPrizeSnapshot(e, game.ID)
	if !ok {
		return
	}
	if prizeIndex(snap.prizes, values[0]) < 0 {
		helpers.RespondEphemeralCard(e, ui.Error("Ce lot n'existe plus."))
		return
	}
	updateMenu(e, prizeMenuCard(snap, values[0], ""))
}

func HandlePrizeEditButton(e *events.ComponentInteractionCreate, args []string) {
	game, prizeID, ok := requirePrizeArgs(e, args)
	if !ok {
		return
	}

	prize, err := prizeByID(context.Background(), game.ID, prizeID)
	if err != nil {
		respondPrizeError(e, game.ID, err)
		return
	}

	openPrizeModal(e, router.BuildCustomID(modalPrizeEdit, game.ID, prize.ID), "Modifier un lot",
		prize.Label, strconv.Itoa(prize.Position+1))
}

func HandlePrizeDeleteButton(e *events.ComponentInteractionCreate, args []string) {
	game, prizeID, ok := requirePrizeArgs(e, args)
	if !ok {
		return
	}

	ctx := context.Background()
	prize, err := prizeByID(ctx, game.ID, prizeID)
	if err != nil {
		respondPrizeError(e, game.ID, err)
		return
	}
	if err := deletePrize(ctx, game.ID, prizeID); err != nil {
		respondPrizeError(e, game.ID, err)
		return
	}

	snap, ok := loadPrizeSnapshot(e, game.ID)
	if !ok {
		return
	}
	editPublicMessage(e.Client(), snap)
	updateMenu(e, prizeMenuCard(snap, "", fmt.Sprintf("🗑️ Le lot **%s** a été supprimé.", prize.Label)))
}

func HandlePrizeAddModal(e *events.ModalSubmitInteractionCreate, args []string) {
	game, ok := requireActiveGame(e, args)
	if !ok {
		return
	}
	if !requireAdmin(e, e.Member()) {
		return
	}

	label, rank, ok := parsePrizeInputs(e)
	if !ok {
		return
	}

	if err := addPrize(context.Background(), game.ID, label, rank); err != nil {
		respondPrizeError(e, game.ID, err)
		return
	}
	finishPrizeModal(e, game.ID, fmt.Sprintf("✅ Le lot **%s** a été ajouté.", label))
}

func HandlePrizeEditModal(e *events.ModalSubmitInteractionCreate, args []string) {
	game, prizeID, ok := requirePrizeArgs(e, args)
	if !ok {
		return
	}

	label, rank, ok := parsePrizeInputs(e)
	if !ok {
		return
	}

	if err := updatePrize(context.Background(), game.ID, prizeID, label, rank); err != nil {
		respondPrizeError(e, game.ID, err)
		return
	}
	finishPrizeModal(e, game.ID, fmt.Sprintf("✏️ Le lot **%s** a été modifié.", label))
}

type adminResponder interface {
	helpers.EphemeralResponder
	Member() *discord.ResolvedMember
}

func requirePrizeArgs(e adminResponder, args []string) (*ent.LotoGame, string, bool) {
	if len(args) < 2 || args[1] == "" {
		helpers.RespondEphemeralCard(e, ui.Error("Lot introuvable."))
		return nil, "", false
	}
	game, ok := requireActiveGame(e, args)
	if !ok {
		return nil, "", false
	}
	if !requireAdmin(e, e.Member()) {
		return nil, "", false
	}
	return game, args[1], true
}

func openPrizeModal(e *events.ComponentInteractionCreate, customID, title, label, rank string) {
	labelInput := discord.NewParagraphTextInput(inputPrizeLabel).
		WithRequired(true).
		WithMaxLength(maxPrizeLabelLen).
		WithPlaceholder("Ex: Une voiture de sport")
	if label != "" {
		labelInput = labelInput.WithValue(label)
	}

	rankInput := discord.NewShortTextInput(inputPrizeRank).
		WithRequired(false).
		WithMaxLength(2).
		WithPlaceholder("Vide = dernière position")
	if rank != "" {
		rankInput = rankInput.WithValue(rank)
	}

	modal := discord.NewModalCreate(customID, title).
		AddLabel("Intitulé du lot", labelInput).
		AddLabel("Position dans le tirage (1 = premier tiré)", rankInput)

	if err := e.Modal(modal); err != nil {
		logger.Error("loto: opening prize modal", "error", err)
	}
}

func parsePrizeInputs(e *events.ModalSubmitInteractionCreate) (string, int, bool) {
	label := strings.TrimSpace(e.Data.Text(inputPrizeLabel))
	if label == "" {
		helpers.RespondEphemeralCard(e, ui.Error("L'intitulé du lot ne peut pas être vide."))
		return "", 0, false
	}
	if len(label) > maxPrizeLabelLen {
		helpers.RespondEphemeralCard(e, ui.Error(fmt.Sprintf(msgTooLong, "L'intitulé du lot", maxPrizeLabelLen)))
		return "", 0, false
	}

	raw := strings.TrimSpace(e.Data.Text(inputPrizeRank))
	if raw == "" {
		return label, 0, true
	}
	rank, err := strconv.Atoi(raw)
	if err != nil || rank <= 0 {
		helpers.RespondEphemeralCard(e, ui.Error("La position doit être un entier strictement positif."))
		return "", 0, false
	}
	return label, rank, true
}

func finishPrizeModal(e *events.ModalSubmitInteractionCreate, gameID, notice string) {
	snap, ok := loadPrizeSnapshot(e, gameID)
	if !ok {
		return
	}
	editPublicMessage(e.Client(), snap)

	if err := e.UpdateMessage(prizeMenuCard(snap, "", notice).MessageUpdate()); err != nil {
		logger.Error("loto: updating prize menu", "game", gameID, "error", err)
	}
}

func updateMenu(e *events.ComponentInteractionCreate, card *ui.Card) {
	if err := e.UpdateMessage(card.MessageUpdate()); err != nil {
		logger.Error("loto: updating prize menu", "error", err)
	}
}

func loadPrizeSnapshot(r helpers.EphemeralResponder, gameID string) (*snapshot, bool) {
	snap, err := loadSnapshot(context.Background(), gameID)
	if err != nil {
		logger.Error("loto: loading game for prizes", "game", gameID, "error", err)
		helpers.RespondEphemeralCard(r, ui.Error("Erreur lors de la lecture du loto."))
		return nil, false
	}
	if snap == nil {
		helpers.RespondEphemeralCard(r, ui.Error("Ce loto n'existe plus."))
		return nil, false
	}
	return snap, true
}

func respondPrizeError(r helpers.EphemeralResponder, gameID string, err error) {
	switch {
	case errors.Is(err, errPrizeNotFound):
		helpers.RespondEphemeralCard(r, ui.Error("Ce lot n'existe plus."))
	case errors.Is(err, errGameClosed):
		helpers.RespondEphemeralCard(r, ui.Error("Ce loto est terminé."))
	case errors.Is(err, errPrizeLimit):
		helpers.RespondEphemeralCard(r, ui.Error(fmt.Sprintf("Un loto ne peut pas avoir plus de %d lots.", maxPrizeCount)))
	case errors.Is(err, errLastPrize):
		helpers.RespondEphemeralCard(r, ui.Error("Impossible de supprimer le dernier lot : un loto doit en avoir au moins un."))
	default:
		logger.Error("loto: managing prizes", "game", gameID, "error", err)
		helpers.RespondEphemeralCard(r, ui.Error("Erreur lors de la mise à jour des lots."))
	}
}

func prizeMenuCard(s *snapshot, selectedID, notice string) *ui.Card {
	card := ui.New().
		Accent(lotoColor).
		Titlef("🎁 Lots du loto : %s", s.game.Name)

	if notice != "" {
		card.Text(notice)
	}

	if len(s.prizes) == 0 {
		card.Text("Aucun lot configuré.")
	} else {
		for _, chunk := range chunkLines(prizeLines(s), maxSectionLength) {
			card.Text(chunk)
		}
	}

	card.Divider()
	for _, menu := range prizeSelects(s, selectedID) {
		card.Row(menu)
	}

	if index := prizeIndex(s.prizes, selectedID); index >= 0 {
		prize := s.prizes[index]
		card.Divider().
			Textf("Lot sélectionné : **#%d** %s", index+1, prize.Label).
			Row(
				discord.NewPrimaryButton("Modifier", router.BuildCustomID(buttonPrizeEdit, s.game.ID, prize.ID)).
					WithEmoji(discord.NewComponentEmoji("✏️")),
				discord.NewDangerButton("Supprimer", router.BuildCustomID(buttonPrizeDelete, s.game.ID, prize.ID)).
					WithEmoji(discord.NewComponentEmoji("🗑️")).
					WithDisabled(len(s.prizes) <= 1),
			)
	}

	return card.Subtextf("%d/%d lots · Choisissez un lot à modifier ou ajoutez-en un.", len(s.prizes), maxPrizeCount)
}

func prizeSelects(s *snapshot, selectedID string) []discord.InteractiveComponent {
	options := make([]discord.StringSelectMenuOption, 0, len(s.prizes)+1)
	for i, prize := range s.prizes {
		options = append(options, discord.NewStringSelectMenuOption(
			truncate(fmt.Sprintf("#%d %s", i+1, prize.Label), maxSelectLabelLength), prize.ID).
			WithEmoji(discord.NewComponentEmoji("🎁")).
			WithDefault(prize.ID == selectedID))
	}
	if len(s.prizes) < maxPrizeCount {
		options = append(options, discord.NewStringSelectMenuOption("Ajouter un lot", prizeAddValue).
			WithEmoji(discord.NewComponentEmoji("➕")))
	}

	menus := make([]discord.InteractiveComponent, 0, len(options)/maxSelectOptions+1)
	for start := 0; start < len(options); start += maxSelectOptions {
		end := min(start+maxSelectOptions, len(options))
		menus = append(menus, discord.NewStringSelectMenu(
			router.BuildCustomID(selectPrize, s.game.ID, strconv.Itoa(len(menus))),
			"Choisir un lot…", options[start:end]...))
	}
	return menus
}
