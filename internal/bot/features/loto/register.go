package loto

import "Eve/internal/bot/router"

const (
	buttonBuy         = "loto:buy"
	buttonRemove      = "loto:remove"
	buttonEditPlayer  = "loto:editplayer"
	buttonDraw        = "loto:draw"
	buttonDrawConfirm = "loto:drawconfirm"
	buttonPrizes      = "loto:prizes"
	buttonPrizeEdit   = "loto:prizeedit"
	buttonPrizeDelete = "loto:prizedel"

	selectPrize = "loto:prizepick"

	modalBuy        = "loto:buymodal"
	modalRemove     = "loto:removemodal"
	modalEditPlayer = "loto:editplayermodal"
	modalPrizeAdd   = "loto:prizeaddmodal"
	modalPrizeEdit  = "loto:prizeeditmodal"
)

const prizeAddValue = "add"

const (
	inputPlayerName  = "playerName"
	inputTicketCount = "count"
	inputOldName     = "oldName"
	inputNewName     = "newName"
	inputPrizeLabel  = "prizeLabel"
	inputPrizeRank   = "prizeRank"
)

func Register(r *router.Router) {
	r.OnCommand("loto", HandleCommand)

	r.OnButton(buttonBuy, HandleBuyButton)
	r.OnButton(buttonRemove, HandleRemoveButton)
	r.OnButton(buttonEditPlayer, HandleEditPlayerButton)
	r.OnButton(buttonDraw, HandleDrawButton)
	r.OnButton(buttonDrawConfirm, HandleDrawConfirmButton)
	r.OnButton(buttonPrizes, HandlePrizesButton)
	r.OnButton(buttonPrizeEdit, HandlePrizeEditButton)
	r.OnButton(buttonPrizeDelete, HandlePrizeDeleteButton)

	r.OnSelectMenu(selectPrize, HandlePrizeSelect)

	r.OnModal(modalBuy, HandleBuyModal)
	r.OnModal(modalRemove, HandleRemoveModal)
	r.OnModal(modalEditPlayer, HandleEditPlayerModal)
	r.OnModal(modalPrizeAdd, HandlePrizeAddModal)
	r.OnModal(modalPrizeEdit, HandlePrizeEditModal)
}
