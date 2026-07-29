package blague

import "Eve/internal/bot/router"

// Register wires the /blague command and its "Rendre publique" button.
//
// Handlers are registered unconditionally, even without a token: the command is
// not published to Discord in that case, but a stale registration on Discord's
// side would otherwise hit the router's "interaction inconnue" path instead of
// the explicit "fonctionnalité désactivée" answer.
func Register(r *router.Router) {
	r.OnCommand(CommandName, HandleCommand)
	r.OnButton(CustomIDPublic, HandlePublicButton)
}
