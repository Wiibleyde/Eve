package music

import "testing"

func TestSplitArtistTitle(t *testing.T) {
	cases := []struct {
		raw    string
		artist string
		title  string
	}{
		{`Clair Obscur: Expedition 33 | Lumière [Official Music Video]`, "Clair Obscur: Expedition 33", "Lumière"},
		{"Imagine Dragons - Warriors (Official Lyric Video)", "Imagine Dragons", "Warriors"},
		{"Warriors", "", "Warriors"},
	}

	for _, c := range cases {
		artist, title := splitArtistTitle(c.raw)
		if artist != c.artist || title != c.title {
			t.Errorf("splitArtistTitle(%q) = (%q, %q), want (%q, %q)", c.raw, artist, title, c.artist, c.title)
		}
	}
}

func TestContainmentRejectsDifferentSong(t *testing.T) {
	if score := containment("Lumière", "Une vie à t'aimer"); score >= lyricsMatchThreshold {
		t.Errorf(`containment("Lumière", "Une vie à t'aimer") = %v, want below %v`, score, lyricsMatchThreshold)
	}
	if score := containment("Lumière", "Nos vies en Lumière"); score < lyricsMatchThreshold {
		t.Errorf(`containment("Lumière", "Nos vies en Lumière") = %v, want at least %v`, score, lyricsMatchThreshold)
	}
}

func TestContainmentAcceptsMatchingTitles(t *testing.T) {
	cases := [][2]string{
		{"Lumière", "Lumière"},
		{"Lumière", `Lumière (From "Clair Obscur: Expedition 33")`},
		{"Warriors", "Warriors"},
		{"Warriors", "Warriors (from “League of Legends”) [Live]"},
	}

	for _, c := range cases {
		if score := containment(c[0], c[1]); score < lyricsMatchThreshold {
			t.Errorf("containment(%q, %q) = %v, want at least %v", c[0], c[1], score, lyricsMatchThreshold)
		}
	}
}

func TestCleanArtist(t *testing.T) {
	if got := cleanArtist("Imagine Dragons - Topic"); got != "Imagine Dragons" {
		t.Errorf(`cleanArtist("Imagine Dragons - Topic") = %q, want "Imagine Dragons"`, got)
	}
}
