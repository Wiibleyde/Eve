package game

import (
	"encoding/json"
	"io"
	"main/pkg/logger"
	"net/http"
	"strings"
	"unicode"
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

type MotusGame struct {
	GameState   MotusGameState
	Word        string
	Tries       []string
	Corrections [][]MotusLetterState
}

func NewMotusGame(word string) *MotusGame {
	return &MotusGame{
		GameState:   MotusGameStatePlaying,
		Word:        normalize(word),
		Tries:       []string{},
		Corrections: [][]MotusLetterState{},
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

func (g *MotusGame) HandleTry(try string) {
	if g.GameState != MotusGameStatePlaying {
		return
	}

	try = normalize(try)
	g.Word = normalize(g.Word)

	g.Tries = append(g.Tries, try)
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
	g.Corrections = append(g.Corrections, correction)

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
