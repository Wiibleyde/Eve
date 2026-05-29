package helpers

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

// RequirePermission checks whether the caller has the given permission.
// Returns false and sends an ephemeral error if the permission is missing.
func RequirePermission(e *events.ApplicationCommandInteractionCreate, perm discord.Permissions, msg string) bool {
	member := e.Member()
	if member == nil || !member.Permissions.Has(perm) {
		RespondEphemeral(e, msg)
		return false
	}
	return true
}
