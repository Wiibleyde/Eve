package calendar

import "Eve/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand("calendar", HandleCommand)
}
