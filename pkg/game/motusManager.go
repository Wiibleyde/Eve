package game

import (
	"encoding/json"
	"io"
	"main/pkg/bot_utils"
	"main/pkg/logger"
	"net/http"
	"strings"
	"unicode"

	"fmt"

	"github.com/bwmarrin/discordgo"
)

const (
	motusMaxAttempts = 6
)

type MotusGameState int

const (
	MotusGameStatePlaying MotusGameState = iota
	MotusGameStateWon
	MotusGameStateLost
)

type MotusLetterState int

const (
	MotusLetterStateFound MotusLetterState = iota
	MotusLetterStateMisplaced
	MotusLetterStateNotFound
	MotusLetterStateEmpty
)

type MotusTry struct {
	Username   string
	Try        string
	Correction []MotusLetterState
}

type MotusGame struct {
	GameState MotusGameState
	Word      string
	Tries     []MotusTry
}

type MotusGameManager struct {
	Games map[string]*MotusGame
}

var motusGameManager MotusGameManager

func NewMotusGame() *MotusGame {
	word := GetRandomWord()
	logger.DebugLogger.Println("New motus game with word:", word)
	return &MotusGame{
		GameState: MotusGameStatePlaying,
		Word:      normalize(word),
		Tries:     []MotusTry{},
	}
}

func normalize(input string) string {
	var normalized strings.Builder
	for _, r := range strings.ToUpper(input) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		normalized.WriteRune(r)
	}
	return normalized.String()
}

func (g *MotusGame) HandleTry(try string, username string) {
	if g.GameState != MotusGameStatePlaying {
		return
	}

	try = normalize(try)
	g.Word = normalize(g.Word)

	correction := make([]MotusLetterState, len(try))
	for i, letter := range try {
		if letter == rune(g.Word[i]) {
			correction[i] = MotusLetterStateFound
		} else if g.WordContains(letter) {
			correction[i] = MotusLetterStateMisplaced
		} else {
			correction[i] = MotusLetterStateNotFound
		}
	}
	g.Tries = append(g.Tries, MotusTry{
		Username:   username,
		Try:        try,
		Correction: correction,
	})

	if try == g.Word {
		g.GameState = MotusGameStateWon
	} else if len(g.Tries) >= motusMaxAttempts {
		g.GameState = MotusGameStateLost
	}
}

func (g *MotusGame) WordContains(letter rune) bool {
	for _, l := range g.Word {
		if l == letter {
			return true
		}
	}
	return false
}

func (g *MotusGame) GetEmbed(s *discordgo.Session) *discordgo.MessageEmbed {
	embed := bot_utils.BasicEmbedBuilder(s)

	color := 0x00ff00
	switch g.GameState {
	case MotusGameStateWon:
		color = 0x00ff00
	case MotusGameStateLost:
		color = 0xff0000
	case MotusGameStatePlaying:
		if len(g.Tries) <= 2 {
			color = 0x00ff00
		} else if len(g.Tries) <= 4 {
			color = 0xffff00
		} else {
			color = 0xff0000
		}
	}
	embed.Color = color

	embed.Title = "Motus"

	if g.GameState == MotusGameStatePlaying {
		embed.Description = "Essaye de deviner le mot ! \n(Mot en + " + fmt.Sprint(len(g.Word)) + " lettres)"
	} else if g.GameState == MotusGameStateWon {
		embed.Description = "Bravo ! Vous avez trouvé le mot !"
	} else if g.GameState == MotusGameStateLost {
		embed.Description = "Dommage ! Tu as perdu ! Le mot était : " + g.Word
	}

	embed.Fields = make([]*discordgo.MessageEmbedField, 0, len(g.Tries)+1)
	for i, motusTry := range g.Tries {
		correctionStr := ""
		for _, state := range motusTry.Correction {
			switch state {
			case MotusLetterStateFound:
				correctionStr += "🟩"
			case MotusLetterStateMisplaced:
				correctionStr += "🟨"
			case MotusLetterStateNotFound:
				correctionStr += "🟥"
			case MotusLetterStateEmpty:
				correctionStr += "⬜"
			}
		}
		tryStr := ""
		for _, letter := range motusTry.Try {
			tryStr += " " + string(letter) + " "
		}
		tryStr = "`" + tryStr + "`"
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Essai " + fmt.Sprint(i+1) + " : " + motusTry.Username,
			Value: tryStr + "\n" + correctionStr,
		})
	}

	return embed
}

func (g *MotusGame) GetComponents() []discordgo.MessageComponent {
	button := &discordgo.Button{
		Label:    "Jouer un essai",
		Style:    discordgo.PrimaryButton,
		CustomID: "motusTryButton",
	}

	actionRow := &discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			button,
		},
	}

	return []discordgo.MessageComponent{
		actionRow,
	}
}

func GetRandomWord() string {
	response, err := http.Get("https://trouve-mot.fr/api/random")
	if err != nil {
		logger.ErrorLogger.Println("Error fetching random word:", err)
		return ""
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logger.ErrorLogger.Println("Error fetching random word: status code", response.StatusCode)
		return ""
	}
	result, err := io.ReadAll(response.Body)
	if err != nil {
		logger.ErrorLogger.Println("Error reading response body:", err)
		return ""
	}
	var data []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(result, &data); err != nil {
		logger.ErrorLogger.Println("Error unmarshalling JSON:", err)
		return ""
	}
	if len(data) == 0 {
		logger.ErrorLogger.Println("No words found in response")
		return ""
	}
	word := data[0].Name
	normalizedWord := normalize(word)
	return normalizedWord
}

func GetMotusGame(messageID string) *MotusGame {
	if game, ok := motusGameManager.Games[messageID]; ok {
		return game
	}
	return nil
}

func AddMotusGame(messageID string, game *MotusGame) {
	if motusGameManager.Games == nil {
		motusGameManager.Games = make(map[string]*MotusGame)
	}
	motusGameManager.Games[messageID] = game
}

func RemoveMotusGame(messageID string) {
	delete(motusGameManager.Games, messageID)
}
