package config

import (
	"fmt"
	"strconv"
	"strings"

	"Eve/internal/database/tables"
)

type Kind int

const (
	KindChannel Kind = iota
	KindRole
	KindBool
	KindString
)

type Key struct {
	Name        string
	Command     string
	Description string
	Kind        Kind
}

var Keys = []Key{
	{
		Name:        tables.BirthdayChannel.String(),
		Command:     "birthday-channel",
		Description: "Salon des anniversaires",
		Kind:        KindChannel,
	},
	{
		Name:        tables.QuoteChannel.String(),
		Command:     "quote-channel",
		Description: "Salon des citations",
		Kind:        KindChannel,
	},
	{
		Name:        tables.JokesDisabled.String(),
		Command:     "jokes-disabled",
		Description: "Réponses automatiques (blagues) désactivées",
		Kind:        KindBool,
	},
}

func keyByCommand(command string) (Key, bool) {
	for _, k := range Keys {
		if k.Command == command {
			return k, true
		}
	}
	return Key{}, false
}

func keyByName(name string) (Key, bool) {
	for _, k := range Keys {
		if k.Name == name {
			return k, true
		}
	}
	return Key{}, false
}

func keyNames() []string {
	names := make([]string, 0, len(Keys))
	for _, k := range Keys {
		names = append(names, k.Name)
	}
	return names
}

func formatValue(k Key, raw string) string {
	switch k.Kind {
	case KindChannel:
		return "<#" + raw + ">"
	case KindRole:
		return "<@&" + raw + ">"
	case KindBool:
		if parseBool(raw) {
			return "Oui"
		}
		return "Non"
	default:
		return "`" + raw + "`"
	}
}

func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "oui", "yes", "on":
		return true
	default:
		return false
	}
}

func FormatBool(v bool) string { return strconv.FormatBool(v) }

func choiceLabel(k Key) string {
	return truncate(fmt.Sprintf("%s (%s)", k.Description, k.Name), maxChoiceNameLength)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return string(runes[:n])
	}
	return string(runes[:n-1]) + "…"
}
