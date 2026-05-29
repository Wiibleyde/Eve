package config

import "Eve/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand("config", HandleCommand)
}
