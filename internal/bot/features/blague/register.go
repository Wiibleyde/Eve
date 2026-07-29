package blague

import "Eve/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand(CommandName, HandleCommand)
	r.OnButton(CustomIDPublic, HandlePublicButton)
}
