package streamer

import (
	"strings"
	"testing"
	"time"

	"Eve/internal/twitch"
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

	// A corrupted row must not turn into a raw, unrestricted mention.
	content, mentions = roleMention("not-a-snowflake")
	if content != "" || len(mentions.Roles) != 0 {
		t.Fatalf("invalid role must ping nothing, got %q %+v", content, mentions)
	}
}

func TestNoMentionsParsesNothing(t *testing.T) {
	// A nil Parse marshals as JSON null, which Discord may read as "default
	// parsing"; only an explicit empty array means "no mention type at all".
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
	if len(got) > embedDescriptionLimit {
		t.Fatalf("joinCapped produced %d bytes, over the %d limit", len(got), embedDescriptionLimit)
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
	got := truncate(long, embedTitleLimit)
	if len([]rune(got)) != embedTitleLimit {
		t.Fatalf("truncate produced %d runes, want %d", len([]rune(got)), embedTitleLimit)
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

	embed := liveEmbed(s, twitch.User{}, false)
	if embed.URL != "https://twitch.tv/someone" {
		t.Fatalf("embed URL = %q", embed.URL)
	}
	if embed.Image == nil || !strings.Contains(embed.Image.URL, "1280x720.jpg") {
		t.Fatalf("thumbnail placeholders were not filled: %+v", embed.Image)
	}
	if len(embed.Fields) != 3 {
		t.Fatalf("expected game/viewers/uptime fields, got %d", len(embed.Fields))
	}
	if embed.Fields[1].Value != "1234" {
		t.Fatalf("viewer field = %q", embed.Fields[1].Value)
	}
}

func TestEndedEmbedWithoutObservedStart(t *testing.T) {
	// After a restart the duration is unknown; the embed must still render.
	embed := endedEmbed(&trackState{}, "someone", twitch.User{}, false)
	if embed.Title != "Stream terminé" {
		t.Fatalf("title = %q", embed.Title)
	}
	for _, f := range embed.Fields {
		if f.Name == "Durée" {
			t.Fatal("duration must be omitted when the start was never observed")
		}
	}

	embed = endedEmbed(&trackState{seen: true, title: "T", game: "G", startedAt: time.Now().Add(-90 * time.Minute)}, "someone", twitch.User{}, false)
	var hasDuration bool
	for _, f := range embed.Fields {
		if f.Name == "Durée" {
			hasDuration = true
		}
	}
	if !hasDuration {
		t.Fatal("duration must be shown when the start was observed")
	}
}
