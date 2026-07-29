package talk

import (
	"Eve/internal/bot/helpers"
	"Eve/internal/bot/router"
	"Eve/internal/logger"
)

func Register(r *router.Router) {
	if !helpers.OwnerConfigured() {
		logger.Warn("BOT_OWNER_ID is missing or invalid, /talk is disabled")
		return
	}
	r.OnCommand(CommandName, HandleCommand)
}
