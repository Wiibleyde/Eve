package streamer

import "Eve/internal/bot/router"

func Register(r *router.Router) {
	if !Enabled() {
		warnDisabled()
		return
	}
	r.OnCommand(CommandName, HandleCommand)
}
