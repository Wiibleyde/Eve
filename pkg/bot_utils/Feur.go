package bot_utils

import (
	"math/rand"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type FeurAnswer struct {
	Response    string
	Probability int
}

var possibleFeurResponses = []FeurAnswer{
	{
		Response:    "Feur.",
		Probability: 70,
	},
	{
		Response:    "coubeh.",
		Probability: 10,
	},
	{
		Response:    "la 🐨",
		Probability: 10,
	},
	{
		Response:    "drilatère.",
		Probability: 10,
	},
}

var quoiRegexps = []*regexp.Regexp{
	// Quoi
	regexp.MustCompile(`(?i)qu+o+i+[\ ?]*\?*$`),
	regexp.MustCompile(`(?i)qu+o+i+[\ +]*\?*$`),

	// Koa
	regexp.MustCompile(`(?i)ko+a+[\ ?]*\?*$`),
	regexp.MustCompile(`(?i)ko+a+[\ +]*\?*$`),

	// Qoa
	regexp.MustCompile(`(?i)q+o+a+[\ ?]*\?*$`),
	regexp.MustCompile(`(?i)q+o+a+[\ +]*\?*$`),

	// Koi
	regexp.MustCompile(`(?i)ko+i+[\ ?]*\?*$`),
	regexp.MustCompile(`(?i)ko+i+[\ +]*\?*$`),

	// Kwa
	regexp.MustCompile(`(?i)kw+a+[\ ?]*\?*$`),
	regexp.MustCompile(`(?i)kw+a+[\ +]*\?*$`),

	// Kewa
	regexp.MustCompile(`(?i)k+e+w+a+[\ ?]*\?*$`),
	regexp.MustCompile(`(?i)k+e+w+a+[\ +]*\?*$`),
}

// DetectFeur checks if a message contains a word ending with "quoi" or its variants
// and responds with a randomly selected "Feur" response
func DetectFeur(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from bots including itself
	if m.Author.Bot {
		return
	}

	message := strings.ToLower(m.Content)
	if IsFeurDetected(message) {
		response := GenerateFeurResponse()
		_, _ = s.ChannelMessageSend(m.ChannelID, response)
	}
}

// IsFeurDetected checks if a message contains one of the "quoi" patterns
func IsFeurDetected(message string) bool {
	for _, regex := range quoiRegexps {
		if regex.MatchString(message) {
			return true
		}
	}
	return false
}

// GenerateFeurResponse returns a random response based on the probability distribution
func GenerateFeurResponse() string {
	random := rand.Float64() * 100
	cumulativeProbability := 0

	for _, response := range possibleFeurResponses {
		cumulativeProbability += response.Probability
		if random <= float64(cumulativeProbability) {
			return response.Response
		}
	}

	return "Feur."
}
