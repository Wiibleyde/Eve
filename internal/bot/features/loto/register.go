package loto

import "Eve/internal/bot/router"

// Custom ID prefixes, "<feature>:<action>" per the router scheme. The single
// data segment is the game UUID, keeping every ID well under Discord's 100
// character cap (longest is "loto:editplayermodal:" + 36 = 57).
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

// Modal text input custom IDs.
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
