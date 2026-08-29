package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

const MaxRadioItems = 25

func Radio(ctx context.Context, uri string) ([]Media, error) {
	binary := ytDlp()
	if binary == "" {
		return nil, ErrNoExtractor
	}

	id := VideoID(uri)
	if id == "" {
		return nil, ErrNoMedia
	}

	output, err := run(ctx, binary,
		"--dump-single-json",
		"--flat-playlist",
		"--playlist-end", strconv.Itoa(MaxRadioItems),
		"--no-warnings",
		"--no-progress",
		"--ignore-config",
		"https://www.youtube.com/watch?v="+id+"&list=RD"+id,
	)
	if err != nil {
		return nil, err
	}

	var payload rawEntry
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, fmt.Errorf("parsing yt-dlp radio output: %w", err)
	}

	items := make([]Media, 0, len(payload.Entries))
	for _, entry := range payload.Entries {
		if entry.ID == id {
			continue
		}
		if media, ok := entry.media(); ok {
			items = append(items, media)
		}
	}
	if len(items) == 0 {
		return nil, ErrNoMedia
	}
	return items, nil
}
