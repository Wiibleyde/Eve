package loto

import "Eve/internal/bot/router"

const (
	buttonBuy         = "loto:buy"
	buttonRemove      = "loto:remove"
	buttonEditPlayer  = "loto:editplayer"
	buttonDraw        = "loto:draw"
	buttonDrawConfirm = "loto:drawconfirm"

	modalBuy        = "loto:buymodal"
	modalRemove     = "loto:removemodal"
	modalEditPlayer = "loto:editplayermodal"
)

const (
	inputPlayerName  = "playerName"
	inputTicketCount = "count"
	inputOldName     = "oldName"
	inputNewName     = "newName"
)

func Register(r *router.Router) {
	r.OnCommand("loto", HandleCommand)

	r.OnButton(buttonBuy, HandleBuyButton)
	r.OnButton(buttonRemove, HandleRemoveButton)
	r.OnButton(buttonEditPlayer, HandleEditPlayerButton)
	r.OnButton(buttonDraw, HandleDrawButton)
	r.OnButton(buttonDrawConfirm, HandleDrawConfirmButton)

	r.OnModal(modalBuy, HandleBuyModal)
	r.OnModal(modalRemove, HandleRemoveModal)
	r.OnModal(modalEditPlayer, HandleEditPlayerModal)
}
