package mpthreads

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/disgoorg/disgo/rest"
)

func jsonError(code rest.JSONErrorCode) error {
	return &rest.Error{Code: code, Message: "test"}
}

func TestDeliveryWarningClosedDMs(t *testing.T) {
	got := deliveryWarning(jsonError(rest.JSONErrorCodeCannotSendMessagesToThisUser), 0)
	if got != msgDMClosed {
		t.Fatalf("expected the closed-DM wording, got %q", got)
	}
}

func TestDeliveryWarningOtherFailureIsNotBlamedOnClosedDMs(t *testing.T) {
	for name, err := range map[string]error{
		"network":   errors.New("connection reset"),
		"too large": jsonError(rest.JSONErrorCodeRequestEntityTooLarge),
		"invalid":   jsonError(rest.JSONErrorCodeInvalidFormBody),
	} {
		if got := deliveryWarning(err, 0); got != msgDMError {
			t.Fatalf("%s: expected the generic wording, got %q", name, got)
		}
	}
}

func TestDeliveryWarningPartialDelivery(t *testing.T) {
	got := deliveryWarning(jsonError(rest.JSONErrorCodeCannotSendMessagesToThisUser), 2)
	if got != msgDMPartial {
		t.Fatalf("a message already partly delivered must be reported as such, got %q", got)
	}
}

func TestDeliveryWarningsAllMentionTheUser(t *testing.T) {
	for _, format := range []string{msgDMClosed, msgDMError, msgDMPartial} {
		if !strings.Contains(fmt.Sprintf(format, "<@1>"), "<@1>") {
			t.Fatalf("warning %q drops the user mention", format)
		}
	}
}

func TestLinksOfLabelsExpiringURLs(t *testing.T) {
	links := linksOf([]relayAttachment{{filename: "photo.png", url: "https://cdn.example/photo.png?ex=1"}})
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if !strings.Contains(links[0], "photo.png") || !strings.Contains(links[0], "https://cdn.example/photo.png?ex=1") {
		t.Fatalf("link lost the filename or the URL: %q", links[0])
	}
	if !strings.Contains(links[0], "temporaire") {
		t.Fatalf("staff must be told the link expires: %q", links[0])
	}
}
