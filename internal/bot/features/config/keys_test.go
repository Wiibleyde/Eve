package config

import (
	"testing"

	"github.com/disgoorg/disgo/discord"
)

func TestRegistryInvariants(t *testing.T) {
	reserved := map[string]bool{"get": true, "reset": true, "list": true}
	names := map[string]bool{}
	commands := map[string]bool{}

	for _, k := range Keys {
		if k.Name == "" || k.Command == "" || k.Description == "" {
			t.Fatalf("incomplete registry entry: %+v", k)
		}
		if reserved[k.Command] {
			t.Fatalf("command %q collides with a shared subcommand", k.Command)
		}
		if names[k.Name] {
			t.Fatalf("duplicate key name %q", k.Name)
		}
		if commands[k.Command] {
			t.Fatalf("duplicate subcommand %q", k.Command)
		}
		names[k.Name] = true
		commands[k.Command] = true
	}
}

func TestInternalKeysNotExposed(t *testing.T) {
	for _, internal := range []string{"calendar.url", "calendar.channel", "calendar.message", "debug.role"} {
		if _, ok := keyByName(internal); ok {
			t.Fatalf("internal key %q must not be user-settable", internal)
		}
	}
}

func TestBuildCommandLimits(t *testing.T) {
	cmd := buildCommand()
	if len(cmd.Options) != len(Keys)+3 {
		t.Fatalf("expected %d subcommands, got %d", len(Keys)+3, len(cmd.Options))
	}
	if len(cmd.Options) > 25 {
		t.Fatalf("too many subcommands: %d", len(cmd.Options))
	}
	for _, opt := range cmd.Options {
		sub, ok := opt.(discord.ApplicationCommandOptionSubCommand)
		if !ok {
			t.Fatalf("option %q is not a subcommand", opt.OptionName())
		}
		if l := len([]rune(sub.Description)); l == 0 || l > maxDescriptionLength {
			t.Fatalf("subcommand %q description length %d", sub.Name, l)
		}
		for _, child := range sub.Options {
			if l := len([]rune(child.OptionDescription())); l == 0 || l > maxDescriptionLength {
				t.Fatalf("option %q description length %d", child.OptionName(), l)
			}
		}
	}
}

func TestFormatValue(t *testing.T) {
	cases := []struct {
		kind Kind
		raw  string
		want string
	}{
		{KindChannel, "123", "<#123>"},
		{KindRole, "456", "<@&456>"},
		{KindBool, "true", "Oui"},
		{KindBool, "false", "Non"},
		{KindBool, "", "Non"},
		{KindString, "abc", "`abc`"},
	}
	for _, c := range cases {
		if got := formatValue(Key{Kind: c.kind}, c.raw); got != c.want {
			t.Fatalf("formatValue(%d, %q) = %q, want %q", c.kind, c.raw, got, c.want)
		}
	}
}
