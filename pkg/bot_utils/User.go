package bot_utils

import "github.com/bwmarrin/discordgo"

func FormatUserIDs(userIDs []string) string {
	if len(userIDs) == 0 {
		return "Aucun utilisateur trouvé."
	}

	formattedIDs := ""
	for _, userID := range userIDs {
		formattedIDs += "<@" + userID + "> "
	}

	return formattedIDs
}

func Contains(mentions []*discordgo.User, user string) bool {
	for _, mention := range mentions {
		if mention.ID == user {
			return true
		}
	}
	return false
}
