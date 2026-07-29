package motus

import "Eve/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand(CommandName, HandleCommand)
	r.OnButton(CustomIDTry, HandleTryButton)
	r.OnModal(CustomIDSubmit, HandleSubmitModal)
}
