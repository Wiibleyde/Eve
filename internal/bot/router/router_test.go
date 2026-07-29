package router

import (
	"reflect"
	"testing"

	"github.com/disgoorg/disgo/events"
)

func TestSplitCustomID(t *testing.T) {
	cases := []struct {
		id   string
		key  string
		args []string
		ok   bool
	}{
		{"motus:try", "motus:try", []string{}, true},
		{"quiz:answer:0:abc-123", "quiz:answer", []string{"0", "abc-123"}, true},
		{"loto:buy:uuid", "loto:buy", []string{"uuid"}, true},
		{"legacy_button", "", nil, false},
		{"", "", nil, false},
		{":action", "", nil, false},
		{"feature:", "", nil, false},
	}
	for _, c := range cases {
		key, args, ok := SplitCustomID(c.id)
		if ok != c.ok || key != c.key || (ok && !reflect.DeepEqual(args, c.args)) {
			t.Errorf("SplitCustomID(%q) = (%q, %v, %v), want (%q, %v, %v)", c.id, key, args, ok, c.key, c.args, c.ok)
		}
	}
}

func TestBuildCustomIDRoundTrip(t *testing.T) {
	id := BuildCustomID("quiz:answer", "0", "q-42")
	if id != "quiz:answer:0:q-42" {
		t.Fatalf("got %q", id)
	}
	key, args, ok := SplitCustomID(id)
	if !ok || key != "quiz:answer" || !reflect.DeepEqual(args, []string{"0", "q-42"}) {
		t.Fatalf("round trip failed: %q %v %v", key, args, ok)
	}
	if got := BuildCustomID("motus:try"); got != "motus:try" {
		t.Fatalf("got %q", got)
	}
}

func TestRegistriesAreDistinct(t *testing.T) {
	r := New()
	r.OnCommand("ping", func(_ *events.ApplicationCommandInteractionCreate) {})
	if len(r.commands) != 1 {
		t.Fatal("command not registered")
	}
	r.OnButton("quiz:answer", func(_ *events.ComponentInteractionCreate, _ []string) {})
	r.OnSelectMenu("quiz:pick", func(_ *events.ComponentInteractionCreate, _ []string) {})
	if len(r.components) != 2 {
		t.Fatalf("expected 2 component handlers, got %d", len(r.components))
	}
	if r.components["quiz:answer"].kind != "button" || r.components["quiz:pick"].kind != "select_menu" {
		t.Fatal("component kinds wrong")
	}
	r.OnModal("quiz:submit", func(_ *events.ModalSubmitInteractionCreate, _ []string) {})
	r.OnUserContextMenu("Avatar", func(_ *events.ApplicationCommandInteractionCreate) {})
	r.OnMessageContextMenu("Citer", func(_ *events.ApplicationCommandInteractionCreate) {})
	if len(r.modals) != 1 || len(r.userCtx) != 1 || len(r.messageCtx) != 1 {
		t.Fatal("registries wrong")
	}
}
