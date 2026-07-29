package motus

const (
	MsgGuildOnly    = "Cette commande doit être utilisée dans un serveur."
	MsgNoWord       = "Impossible de trouver un mot pour le moment. Réessayez plus tard."
	MsgBoardFailed  = "Impossible d'afficher la grille dans ce salon."
	MsgStoreFailed  = "Impossible d'enregistrer la partie."
	MsgGameStarted  = "Partie de Motus lancée ! Tout le monde peut proposer un mot."
	MsgUnknownGame  = "Cette partie n'existe plus."
	MsgDBError      = "Erreur lors de l'accès à la partie."
	MsgExpired      = "La partie a expiré. Le mot était **%s**."
	MsgGameOver     = "Cette partie est terminée. Le mot était **%s**."
	MsgInvalidGuess = "Proposition invalide : il faut exactement **%d** lettres, sans espace ni chiffre."
	MsgAlreadyTried = "**%s** a déjà été proposé. Cet essai ne compte pas."
	MsgWon          = "Bravo ! Le mot était **%s**, trouvé en %d/%d essais."
	MsgLost         = "Perdu ! Le mot était **%s**."
)
