package bot_utils

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
