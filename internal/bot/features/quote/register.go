package quote

import "Eve/internal/bot/router"

func Register(r *router.Router) {
	r.OnCommand("quote", HandleCommand)
	r.OnMessageContextMenu(MenuCreateQuote, HandleCreateQuoteMenu)
}
