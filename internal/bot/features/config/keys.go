package config

import (
	"fmt"
	"strconv"
	"strings"

	"Eve/internal/database/tables"
)

// Kind is the value type of a configuration key. It drives the option type of
// the generated setter subcommand and how a stored value is rendered back to
// the user.
type Kind int

const (
	KindChannel Kind = iota
	KindRole
	KindBool
	KindString
)

// Key describes one user-settable guild configuration entry.
//
// Adding a key to Keys is enough: the setter subcommand, the get/reset choices
// and the /config list output are all derived from this registry.
//
// Internal keys (calendar.*, debug.role) are deliberately absent — they are
// managed by their own feature commands and must not be exposed here.
type Key struct {
	Name        string // guild_configs key, e.g. "birthday.channel"
	Command     string // subcommand name, e.g. "birthday-channel"
	Description string // human label, French, reused in list/get output
	Kind        Kind
}

// Keys is the registry of user-visible configuration keys.
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

// keyNames returns every user-visible key name, for bulk queries.
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

// parseBool is tolerant: values may have been written by other features.
func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "oui", "yes", "on":
		return true
	default:
		return false
	}
}

// FormatBool renders a bool the way this feature stores it, so other features
// reading a Bool key can compare consistently.
func FormatBool(v bool) string { return strconv.FormatBool(v) }

func choiceLabel(k Key) string {
	return truncate(fmt.Sprintf("%s (%s)", k.Description, k.Name), maxChoiceNameLength)
}

// truncate caps a string at n characters (runes), Discord limits are counted in
// characters, not bytes.
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
