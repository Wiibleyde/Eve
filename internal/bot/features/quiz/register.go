package quiz

import "Eve/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand(CommandName, HandleCommand)
	r.OnButton(ButtonAnswerPrefix, HandleAnswer)
}
