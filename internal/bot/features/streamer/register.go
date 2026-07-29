package streamer

import "Eve/internal/bot/router"

// Register wires the /streamer command. Without Twitch credentials the command
// is never registered on Discord either, so there is nothing to dispatch.
func Register(r *router.Router) {
	if !Enabled() {
		warnDisabled()
		return
	}
	r.OnCommand(CommandName, HandleCommand)
}
