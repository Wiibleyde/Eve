package userinfo

import "Eve/internal/bot/router"

func Register(r *router.Router) {
	r.OnUserContextMenu(MenuAvatar, HandleAvatar)
	r.OnUserContextMenu(MenuBanner, HandleBanner)
}
