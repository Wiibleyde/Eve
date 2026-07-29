package mpthreads

import (
	"strings"
	"testing"

	"github.com/disgoorg/disgo/discord"
)

func TestSplitContentKeepsShortBodyIntact(t *testing.T) {
	chunks := splitContent("bonjour\nça va ?")
	if len(chunks) != 1 || chunks[0] != "bonjour\nça va ?" {
		t.Fatalf("expected a single untouched chunk, got %#v", chunks)
	}
}

func TestSplitContentEmpty(t *testing.T) {
	if chunks := splitContent("   \n\n"); len(chunks) != 0 {
		t.Fatalf("expected no chunk for a blank body, got %#v", chunks)
	}
}

func TestSplitContentRespectsLimit(t *testing.T) {
	body := strings.Repeat("é", maxMessageLength*2+37)
	chunks := splitContent(body)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	total := 0
	for _, chunk := range chunks {
		runes := []rune(chunk)
		if len(runes) > maxMessageLength {
			t.Fatalf("chunk longer than the Discord limit: %d runes", len(runes))
		}
		total += len(runes)
	}
	if total != maxMessageLength*2+37 {
		t.Fatalf("content lost while splitting: %d runes kept", total)
	}
}

func TestSplitContentPrefersLineBreaks(t *testing.T) {
	body := strings.Repeat("a", maxMessageLength-10) + "\n" + strings.Repeat("b", 50)
	chunks := splitContent(body)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if strings.Contains(chunks[0], "b") {
		t.Fatalf("split did not happen on the line break: %q", chunks[0])
	}
}

func TestTruncateRunesIsRuneSafe(t *testing.T) {
	got := truncateRunes("ééééé", 3)
	if got != "ééé" {
		t.Fatalf("expected 3 runes, got %q", got)
	}
}

func TestThreadNameFallsBackToID(t *testing.T) {
	user := discord.User{ID: 123456789012345678, Username: "  "}
	if got := threadName(user); got != "123456789012345678" {
		t.Fatalf("expected the user id as fallback, got %q", got)
	}
}

func TestThreadNameIsTruncated(t *testing.T) {
	user := discord.User{ID: 1, Username: strings.Repeat("x", 250)}
	if got := []rune(threadName(user)); len(got) != maxThreadName {
		t.Fatalf("expected %d runes, got %d", maxThreadName, len(got))
	}
}

func TestRelayableTypeFiltersSystemMessages(t *testing.T) {
	if !relayableType(discord.MessageTypeDefault) || !relayableType(discord.MessageTypeReply) {
		t.Fatal("default and reply messages must be relayed")
	}
	if relayableType(discord.MessageTypeChannelPinnedMessage) {
		t.Fatal("system messages must not be relayed")
	}
}

func TestStickerLinksUseTheCDN(t *testing.T) {
	links := stickerLinks([]discord.MessageSticker{{ID: 42, Name: "hello", FormatType: discord.StickerFormatTypePNG}})
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if !strings.Contains(links[0], "hello") || !strings.Contains(links[0], "42") {
		t.Fatalf("unexpected sticker link: %q", links[0])
	}
}
