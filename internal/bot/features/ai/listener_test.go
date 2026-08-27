package ai

import (
	"testing"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

func TestPingableKeepsOnlyAllowedTargets(t *testing.T) {
	allowed := map[snowflake.ID]struct{}{otherID: {}}

	got := pingable("Salut <@461807010086780930> et <@222222222222222222>", allowed)

	if len(got) != 1 || got[0] != otherID {
		t.Errorf("pingable() = %v, want [%d]", got, otherID)
	}
}

func TestPingableIsEmptyWithoutMentions(t *testing.T) {
	allowed := map[snowflake.ID]struct{}{otherID: {}}

	if got := pingable("Aucune mention ici.", allowed); len(got) != 0 {
		t.Errorf("pingable() = %v, want no target", got)
	}
}

func TestAllowedTargetsAlwaysIncludesAuthorAndOwner(t *testing.T) {
	allowed := allowedTargets(999, time.Now(), otherID, ownerID, true)

	if _, ok := allowed[otherID]; !ok {
		t.Error("author is missing from the allowed targets")
	}
	if _, ok := allowed[ownerID]; !ok {
		t.Error("owner is missing from the allowed targets")
	}
}

func TestSpeakerPrefixCarriesTheMention(t *testing.T) {
	got := speaker{id: otherID, name: "Wiibleyde"}.prefix()
	want := "Wiibleyde (<@461807010086780930>) : "
	if got != want {
		t.Errorf("prefix() = %q, want %q", got, want)
	}
}
