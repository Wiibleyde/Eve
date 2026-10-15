package streamer

import (
	"errors"
	"net/http"
	"testing"

	"Eve/internal/bot/ui"
	"Eve/internal/database/ent"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

type fakeRest struct {
	rest.Rest
	updateErr error
	deleteErr error
	updated   int
	deleted   int
}

func (f *fakeRest) UpdateMessage(snowflake.ID, snowflake.ID, discord.MessageUpdate, ...rest.RequestOpt) (*discord.Message, error) {
	f.updated++
	return nil, f.updateErr
}

func (f *fakeRest) DeleteMessage(snowflake.ID, snowflake.ID, ...rest.RequestOpt) error {
	f.deleted++
	return f.deleteErr
}

func restError(status int) error {
	return &rest.Error{Response: &http.Response{StatusCode: status}}
}

func TestRetireNotification(t *testing.T) {
	tests := []struct {
		name        string
		updateErr   error
		deleteErr   error
		wantErr     bool
		wantDeleted int
	}{
		{"edited in place", nil, nil, false, 0},
		{"message already gone", restError(http.StatusNotFound), nil, false, 0},
		{"edit refused, deleted instead", restError(http.StatusBadRequest), nil, false, 1},
		{"edit refused, already deleted", restError(http.StatusBadRequest), restError(http.StatusNotFound), false, 1},
		{"both fail, kept for retry", errors.New("timeout"), errors.New("timeout"), true, 1},
	}

	row := &ent.Stream{ChannelID: "123", MessageID: "456", TwitchLogin: "someone"}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeRest{updateErr: tc.updateErr, deleteErr: tc.deleteErr}
			err := retireNotification(fake, row, ui.New().Title("fin"))
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if fake.updated != 1 || fake.deleted != tc.wantDeleted {
				t.Errorf("updated=%d deleted=%d, want 1 and %d", fake.updated, fake.deleted, tc.wantDeleted)
			}
		})
	}
}

func TestRetireNotificationRejectsInvalidIDs(t *testing.T) {
	fake := &fakeRest{}
	if err := retireNotification(fake, &ent.Stream{ChannelID: "x", MessageID: "456"}, ui.New()); err == nil {
		t.Error("expected an error for an invalid channel ID")
	}
	if fake.updated != 0 || fake.deleted != 0 {
		t.Error("no request should be sent with invalid IDs")
	}
}
