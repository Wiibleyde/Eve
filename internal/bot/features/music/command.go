package music

import (
	"Eve/internal/audio"

	"github.com/disgoorg/disgo/discord"
)

const (
	CommandPlay         = "play"
	CommandQueue        = "queue"
	CommandNowPlaying   = "nowplaying"
	CommandSkip         = "skip"
	CommandBack         = "back"
	CommandPause        = "pause"
	CommandResume       = "resume"
	CommandStop         = "stop"
	CommandLoop         = "loop"
	CommandClear        = "clear"
	CommandRemove       = "remove"
	CommandFilter       = "filter"
	CommandSyncedLyrics = "syncedlyrics"
)

const (
	OptionTrack    = "musique"
	OptionAction   = "action"
	OptionPosition = "position"
	OptionFilter   = "filtre"
)

const (
	CustomIDBack      = "music:back"
	CustomIDSkip      = "music:skip"
	CustomIDPlayPause = "music:playpause"
	CustomIDLoop      = "music:loop"
)

const (
	loopValueQueue    = "queue"
	loopValueTrack    = "track"
	loopValueDisabled = "off"
)

const (
	MsgDisabled       = "La fonctionnalité musique est désactivée : aucun nœud Lavalink n'est configuré."
	MsgNodeDown       = "Le serveur audio est injoignable pour le moment. Réessayez plus tard."
	MsgGuildOnly      = "Cette commande n'est utilisable que sur un serveur."
	MsgNoVoice        = "Vous devez être dans un salon vocal."
	MsgNotPlaying     = "Aucune musique n'est en cours de lecture."
	MsgQueueEmpty     = "Il n'y a aucune musique dans la file d'attente."
	MsgQueueEnded     = "Fin de la file d'attente, je quitte le salon vocal."
	MsgNoResult       = "Aucun résultat trouvé."
	MsgNoExtractor    = "L'extracteur audio (yt-dlp) est introuvable sur le serveur."
	MsgLoadFailed     = "Impossible de charger cette musique."
	MsgPlayFailed     = "Impossible de jouer la musique."
	MsgNoPrevious     = "Il n'y a pas de musique précédente."
	MsgAlreadyPaused  = "La musique est déjà en pause."
	MsgAlreadyPlaying = "La musique est déjà en cours de lecture."
	MsgUnknownFilter  = "Filtre inconnu."
	MsgUnknownLoop    = "Mode de boucle inconnu."
	MsgBadPosition    = "Cette musique n'existe pas dans la file d'attente."
	MsgLyricsMissing  = "Aucune parole synchronisée trouvée pour cette musique."
	MsgLyricsNotTimed = "Aucune parole synchronisée disponible pour cette musique."
	MsgLyricsRunning  = "Les paroles synchronisées sont déjà affichées."
	MsgLyricsFailed   = "Impossible d'afficher les paroles synchronisées."
	MsgThreadFailed   = "Impossible de créer le fil des paroles."
)

var minPosition = 1

func Commands() []discord.ApplicationCommandCreate {
	if !Enabled() {
		return nil
	}

	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        CommandPlay,
			Description: "[Musique] Jouer une musique",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        OptionTrack,
					Description: "Nom ou lien de la musique",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        CommandQueue,
			Description: "[Musique] Afficher la file d'attente",
		},
		discord.SlashCommandCreate{
			Name:        CommandNowPlaying,
			Description: "[Musique] Afficher la musique en cours de lecture",
		},
		discord.SlashCommandCreate{
			Name:        CommandSkip,
			Description: "[Musique] Passer à la musique suivante",
		},
		discord.SlashCommandCreate{
			Name:        CommandBack,
			Description: "[Musique] Revenir à la musique précédente",
		},
		discord.SlashCommandCreate{
			Name:        CommandPause,
			Description: "[Musique] Mettre en pause la musique",
		},
		discord.SlashCommandCreate{
			Name:        CommandResume,
			Description: "[Musique] Reprendre la musique",
		},
		discord.SlashCommandCreate{
			Name:        CommandStop,
			Description: "[Musique] Arrêter la musique",
		},
		discord.SlashCommandCreate{
			Name:        CommandLoop,
			Description: "[Musique] Activer/Désactiver la boucle",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        OptionAction,
					Description: "Mode de boucle",
					Required:    true,
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{Name: "File d'attente", Value: loopValueQueue},
						{Name: "Musique", Value: loopValueTrack},
						{Name: "Désactivée", Value: loopValueDisabled},
					},
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        CommandClear,
			Description: "[Musique] Vider la file d'attente",
		},
		discord.SlashCommandCreate{
			Name:        CommandRemove,
			Description: "[Musique] Supprimer une musique de la file d'attente",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        OptionPosition,
					Description: "Position de la musique à supprimer",
					Required:    true,
					MinValue:    &minPosition,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        CommandFilter,
			Description: "[Musique] Activer/Désactiver un filtre",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        OptionFilter,
					Description: "Nom du filtre",
					Required:    true,
					Choices:     filterChoices(),
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        CommandSyncedLyrics,
			Description: "[Musique] Afficher les paroles synchronisées de la musique en cours",
		},
	}
}

func filterChoices() []discord.ApplicationCommandOptionChoiceString {
	available := audio.Filters()
	choices := make([]discord.ApplicationCommandOptionChoiceString, 0, len(available))
	for _, f := range available {
		choices = append(choices, discord.ApplicationCommandOptionChoiceString{Name: f.Label, Value: f.Name})
	}
	return choices
}
