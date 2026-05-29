package ping

import "Eve/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand("ping", HandleCommand)
}
