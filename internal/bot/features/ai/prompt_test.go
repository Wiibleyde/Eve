package ai

import (
	"strings"
	"testing"

	"github.com/disgoorg/snowflake/v2"
)

const (
	selfID  = snowflake.ID(111111111111111111)
	otherID = snowflake.ID(461807010086780930)
	ownerID = snowflake.ID(222222222222222222)
)

func TestCleanPromptKeepsOtherMentions(t *testing.T) {
	got := cleanPrompt("<@111111111111111111>   salut   <@461807010086780930> ça va ?", selfID)
	want := "salut <@461807010086780930> ça va ?"
	if got != want {
		t.Errorf("cleanPrompt() = %q, want %q", got, want)
	}
}

func TestCleanPromptDefangsMassPings(t *testing.T) {
	if got := cleanPrompt("coucou @everyone et @here", selfID); got != "coucou everyone et here" {
		t.Errorf("cleanPrompt() = %q, want mass pings defanged", got)
	}
}

func TestCleanPromptOnlySelfMentionIsEmpty(t *testing.T) {
	if got := cleanPrompt("<@111111111111111111>", selfID); got != "" {
		t.Errorf("cleanPrompt() = %q, want an empty string", got)
	}
}

func TestCleanReplyKeepsUserMentionsAndDropsTheRest(t *testing.T) {
	got := cleanReply("Salut <@461807010086780930> <@111111111111111111> <@&999> @everyone", selfID)
	want := "Salut <@461807010086780930> everyone"
	if got != want {
		t.Errorf("cleanReply() = %q, want %q", got, want)
	}
}

func TestCleanReplyDropsBrokenMentions(t *testing.T) {
	got := cleanReply("Bonjour @ID du compte et <@ID du compte>", selfID)
	if strings.Contains(got, "ID du compte") {
		t.Errorf("cleanReply() = %q, want no placeholder mention left", got)
	}
}

func TestCleanReplyTruncatesAt1024(t *testing.T) {
	got := cleanReply(strings.Repeat("é", maxReplyLength+50), selfID)
	if length := len([]rune(got)); length != maxReplyLength {
		t.Errorf("length = %d, want %d", length, maxReplyLength)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("cleanReply() = %q, want an ellipsis suffix", got)
	}
}

func TestMentionedUsersDeduplicates(t *testing.T) {
	got := mentionedUsers("<@461807010086780930> et <@461807010086780930> et <@!111111111111111111>")
	if len(got) != 2 || got[0] != otherID || got[1] != selfID {
		t.Errorf("mentionedUsers() = %v, want [%d %d]", got, otherID, selfID)
	}
}

func TestSystemInstructionCarriesIdentityAndEmojis(t *testing.T) {
	content := systemInstruction(selfID, ownerID, true)

	for _, want := range []string{mention(selfID), mention(ownerID), EmojiHappy, "1024 caractères"} {
		if !strings.Contains(content, want) {
			t.Errorf("system prompt is missing %q", want)
		}
	}
}

func TestSystemInstructionOmitsUnknownOwner(t *testing.T) {
	content := systemInstruction(selfID, 0, false)

	if strings.Contains(content, "créateur") {
		t.Errorf("system prompt mentions a creator without a configured owner:\n%s", content)
	}
}
