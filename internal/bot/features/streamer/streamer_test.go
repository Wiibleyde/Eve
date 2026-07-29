package streamer

import (
	"strings"
	"testing"
	"time"

	"Eve/internal/bot/ui"
	"Eve/internal/twitch"

	"github.com/disgoorg/disgo/discord"
)

func TestNormalizeLogin(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{"Wiibleyde", "wiibleyde", true},
		{"  wiibleyde  ", "wiibleyde", true},
		{"@wiibleyde", "wiibleyde", true},
		{"https://www.twitch.tv/wiibleyde", "wiibleyde", true},
		{"http://twitch.tv/wiibleyde/videos", "wiibleyde", true},
		{"twitch.tv/wiibleyde?foo=bar", "wiibleyde", true},
		{"m.twitch.tv/wiibleyde", "wiibleyde", true},
		{"with space", "", false},
		{"ab", "", false},
		{"", "", false},
		{strings.Repeat("a", 26), "", false},
		{"bad-dash", "", false},
	}

	for _, tc := range tests {
		got, ok := normalizeLogin(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("normalizeLogin(%q) = %q,%v want %q,%v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestShouldEdit(t *testing.T) {
	now := time.Now()
	base := twitch.Stream{ID: "1", Title: "Titre", GameName: "Jeu", ViewerCount: 100}

	t.Run("resume after restart", func(t *testing.T) {
		if !shouldEdit(&trackState{}, base, now) {
			t.Fatal("an unobserved state must refresh once")
		}
	})

	t.Run("title change", func(t *testing.T) {
		st := &trackState{seen: true, title: "Ancien", game: "Jeu", viewers: 100, lastEdit: now}
		if !shouldEdit(st, base, now) {
			t.Fatal("a title change must edit immediately")
		}
	})

	t.Run("game change", func(t *testing.T) {
		st := &trackState{seen: true, title: "Titre", game: "Autre", viewers: 100, lastEdit: now}
		if !shouldEdit(st, base, now) {
			t.Fatal("a game change must edit immediately")
		}
	})

	t.Run("nothing changed", func(t *testing.T) {
		st := &trackState{seen: true, title: "Titre", game: "Jeu", viewers: 100, lastEdit: now.Add(-time.Hour)}
		if shouldEdit(st, base, now) {
			t.Fatal("an identical payload must not spend an edit")
		}
	})

	t.Run("viewer count throttled", func(t *testing.T) {
		st := &trackState{seen: true, title: "Titre", game: "Jeu", viewers: 90, lastEdit: now.Add(-time.Minute)}
		if shouldEdit(st, base, now) {
			t.Fatal("a viewer-only change within 5min must be skipped")
		}
	})

	t.Run("viewer count after the interval", func(t *testing.T) {
		st := &trackState{seen: true, title: "Titre", game: "Jeu", viewers: 90, lastEdit: now.Add(-ViewerEditInterval)}
		if !shouldEdit(st, base, now) {
			t.Fatal("a viewer-only change past 5min must edit")
		}
	})
}

func TestTrackStateObserveAndReset(t *testing.T) {
	st := &trackState{missed: 1}
	started := time.Now().Add(-time.Hour)
	st.observe(twitch.Stream{ID: "42", Title: "T", GameName: "G", ViewerCount: 7, StartedAt: started})

	if !st.seen || st.streamID != "42" || st.title != "T" || st.game != "G" || st.viewers != 7 || !st.startedAt.Equal(started) {
		t.Fatalf("observe did not capture the stream: %+v", st)
	}
	if st.missed != 0 {
		t.Fatalf("a live observation must clear the missed counter, got %d", st.missed)
	}

	st.reset()
	if st.seen || st.streamID != "" || !st.startedAt.IsZero() || st.missed != 0 {
		t.Fatalf("reset left state behind: %+v", st)
	}
}

func TestRoleMentionRestrictsMentions(t *testing.T) {
	content, mentions := roleMention("123456789012345678")
	if content != "<@&123456789012345678>" {
		t.Fatalf("content = %q", content)
	}
	if len(mentions.Roles) != 1 || mentions.Roles[0].String() != "123456789012345678" {
		t.Fatalf("allowed mentions must list exactly the configured role: %+v", mentions)
	}
	if len(mentions.Parse) != 0 || len(mentions.Users) != 0 {
		t.Fatalf("nothing else may be mentionable: %+v", mentions)
	}

	content, mentions = roleMention("")
	if content != "" {
		t.Fatalf("no role means no ping, got %q", content)
	}
	if len(mentions.Roles) != 0 || len(mentions.Parse) != 0 {
		t.Fatalf("no role means nothing mentionable: %+v", mentions)
	}

	content, mentions = roleMention("not-a-snowflake")
	if content != "" || len(mentions.Roles) != 0 {
		t.Fatalf("invalid role must ping nothing, got %q %+v", content, mentions)
	}
}

func TestNoMentionsParsesNothing(t *testing.T) {
	if parse := noMentions().Parse; parse == nil || len(parse) != 0 {
		t.Fatalf("Parse must be an explicit empty slice, got %#v", parse)
	}
	if _, mentions := roleMention("123456789012345678"); mentions.Parse == nil {
		t.Fatal("a role ping must still parse nothing else")
	}
}

func TestJoinCapped(t *testing.T) {
	short := []string{"a", "b", "c"}
	if got := joinCapped(short); got != "a\nb\nc" {
		t.Fatalf("joinCapped = %q", got)
	}

	long := make([]string, 300)
	for i := range long {
		long[i] = strings.Repeat("x", 50)
	}
	got := joinCapped(long)
	if len(got) > listLimit {
		t.Fatalf("joinCapped produced %d bytes, over the %d limit", len(got), listLimit)
	}
	if !strings.Contains(got, "de plus") {
		t.Fatalf("the overflow must be reported to the user: %q", got[len(got)-40:])
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{2*time.Hour + 13*time.Minute, "2h 13min"},
		{time.Hour + 5*time.Minute, "1h 05min"},
		{13 * time.Minute, "13min"},
		{45 * time.Second, "45s"},
		{-time.Second, "0s"},
	}
	for _, tc := range tests {
		if got := formatDuration(tc.in); got != tc.want {
			t.Errorf("formatDuration(%s) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCacheBust(t *testing.T) {
	if got := cacheBust(""); got != "" {
		t.Fatalf("empty URL must stay empty, got %q", got)
	}
	if got := cacheBust("https://cdn/preview.jpg"); !strings.Contains(got, "?t=") {
		t.Fatalf("expected a ?t= buster, got %q", got)
	}
	if got := cacheBust("https://cdn/preview.jpg?a=b"); !strings.Contains(got, "&t=") {
		t.Fatalf("expected an &t= buster on an URL that already has a query, got %q", got)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("court", 10); got != "court" {
		t.Fatalf("truncate = %q", got)
	}
	long := strings.Repeat("é", 300)
	got := truncate(long, titleLimit)
	if len([]rune(got)) != titleLimit {
		t.Fatalf("truncate produced %d runes, want %d", len([]rune(got)), titleLimit)
	}
}

func TestLiveEmbedNeverPingsThroughContent(t *testing.T) {
	s := twitch.Stream{
		ID:           "1",
		UserLogin:    "someone",
		UserName:     "SomeOne",
		Title:        "@everyone gros stream",
		GameName:     "Just Chatting",
		ViewerCount:  1234,
		StartedAt:    time.Now().Add(-time.Hour),
		Type:         twitch.StreamTypeLive,
		ThumbnailURL: "https://cdn/{width}x{height}.jpg",
	}

	msg := liveCard(s, twitch.User{}, false, "").MessageCreate()
	if msg.AllowedMentions == nil || len(msg.AllowedMentions.Parse) != 0 {
		t.Fatalf("live card must suppress mentions: %+v", msg.AllowedMentions)
	}

	joined := strings.Join(ui.Texts(msg.Components), "\n")
	for _, want := range []string{"https://twitch.tv/someone", "**Spectateurs** · 1234", "**Jeu** · Just Chatting"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("card is missing %q, got:\n%s", want, joined)
		}
	}
	if !hasGalleryURL(msg.Components, "1280x720.jpg") {
		t.Fatal("thumbnail placeholders were not filled")
	}
}

func hasGalleryURL(components []discord.LayoutComponent, fragment string) bool {
	for _, component := range components {
		container, ok := component.(discord.ContainerComponent)
		if !ok {
			continue
		}
		for sub := range container.SubComponents() {
			gallery, ok := sub.(discord.MediaGalleryComponent)
			if !ok {
				continue
			}
			for _, item := range gallery.Items {
				if strings.Contains(item.Media.URL, fragment) {
					return true
				}
			}
		}
	}
	return false
}

func TestEndedCardWithoutObservedStart(t *testing.T) {
	joined := strings.Join(ui.Texts(endedCard(&trackState{}, "someone", twitch.User{}, false).Components()), "\n")
	if !strings.Contains(joined, "Stream terminé") {
		t.Fatalf("title missing: %q", joined)
	}
	if strings.Contains(joined, "**Durée**") {
		t.Fatal("duration must be omitted when the start was never observed")
	}

	st := &trackState{seen: true, title: "T", game: "G", startedAt: time.Now().Add(-90 * time.Minute)}
	joined = strings.Join(ui.Texts(endedCard(st, "someone", twitch.User{}, false).Components()), "\n")
	if !strings.Contains(joined, "**Durée**") {
		t.Fatal("duration must be shown when the start was observed")
	}
}
