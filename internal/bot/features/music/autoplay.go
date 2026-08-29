package music

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"Eve/internal/audio"
	"Eve/internal/bot/ui"
	"Eve/internal/gemini"
	"Eve/internal/logger"

	"github.com/disgoorg/snowflake/v2"
)

const (
	autoplayPicks     = 5
	autoplayHistory   = 10
	autoplayTimeout   = 2 * time.Minute
	autoplayThreshold = 1
)

const autoplayInstruction = `Tu es un moteur de recommandation musicale. À partir d'une musique de référence et de l'historique d'écoute, propose des musiques du même univers (genre, époque, ambiance).
Règles : ne propose jamais une musique déjà écoutée, varie les artistes, reste sur des morceaux réellement existants et populaires.
Réponds uniquement avec un tableau JSON de la forme [{"artiste":"...","titre":"..."}] sans aucun autre texte.`

func trackKey(media audio.Media) string {
	if id := audio.VideoID(media.URI); id != "" {
		return "yt:" + id
	}
	return strings.ToLower(strings.TrimSpace(media.Author + " - " + media.Title))
}

func knownKeys(tracks []audio.Media) map[string]bool {
	keys := make(map[string]bool, len(tracks))
	for _, track := range tracks {
		keys[trackKey(track)] = true
	}
	return keys
}

func autoplaySeed(state *guildState) (audio.Media, bool) {
	if current := state.current(); current != nil {
		return *current, true
	}
	history := state.recentHistory(1)
	if len(history) == 0 {
		return audio.Media{}, false
	}
	return history[0], true
}

func toggleAutoplay(guildID snowflake.ID) bool {
	state := stateFor(guildID)
	enabled := state.toggleAutoplay()

	refreshNowPlaying(guildID)
	if enabled && state.size() < autoplayThreshold {
		go refillAutoplay(guildID)
	}
	return enabled
}

func refillAutoplay(guildID snowflake.ID) {
	defer recoverAutoplay(guildID)

	state := stateFor(guildID)
	if !state.beginAutoplay() {
		return
	}
	defer state.endAutoplay()

	if state.size() >= autoplayThreshold {
		return
	}

	seed, ok := autoplaySeed(state)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), autoplayTimeout)
	defer cancel()

	picks := suggest(ctx, state, seed)
	if len(picks) == 0 {
		logger.Warn("Music: autoplay found nothing", "guild", guildID.String(), "seed", seed.Title)
		notify(guildID, ui.Warning("Mode automatique", MsgAutoplayEmpty))
		return
	}

	state.enqueue(picks...)
	logger.Info("Music: autoplay queued tracks", "guild", guildID.String(), "seed", seed.Title, "count", len(picks))

	if state.current() != nil {
		refreshNowPlaying(guildID)
		return
	}
	state.cancelLeave()
	playNext(guildID)
}

func recoverAutoplay(guildID snowflake.ID) {
	if r := recover(); r != nil {
		logger.Error("Music: autoplay panicked", "guild", guildID.String(), "panic", fmt.Sprint(r))
	}
}

func suggest(ctx context.Context, state *guildState, seed audio.Media) []audio.Media {
	if !audio.YtDlpAvailable() {
		return nil
	}

	known := knownKeys(state.knownTracks())
	if picks := geminiPicks(ctx, seed, state.recentHistory(autoplayHistory), known); len(picks) > 0 {
		return picks
	}
	return radioPicks(ctx, seed, known)
}

func geminiPicks(ctx context.Context, seed audio.Media, history []audio.Media, known map[string]bool) []audio.Media {
	client := gemini.Default()
	if client == nil {
		return nil
	}

	answer, err := client.CompleteJSON(ctx, autoplayInstruction, autoplayPrompt(seed, history))
	if err != nil {
		logger.Warn("Music: autoplay gemini request failed", "seed", seed.Title, "error", err)
		return nil
	}

	queries := parseSuggestions(answer)
	if len(queries) == 0 {
		logger.Warn("Music: autoplay gemini returned no usable suggestion", "seed", seed.Title)
		return nil
	}
	return resolveAll(ctx, queries, known)
}

func autoplayPrompt(seed audio.Media, history []audio.Media) string {
	var builder strings.Builder
	builder.WriteString("Musique de référence : ")
	builder.WriteString(trackLabel(seed))
	builder.WriteString("\n")

	if len(history) > 0 {
		builder.WriteString("Déjà écouté :\n")
		for _, track := range history {
			builder.WriteString("- ")
			builder.WriteString(trackLabel(track))
			builder.WriteString("\n")
		}
	}

	fmt.Fprintf(&builder, "Propose %d musiques différentes à écouter ensuite.", autoplayPicks)
	return builder.String()
}

func trackLabel(media audio.Media) string {
	if media.Author == "" {
		return media.Title
	}
	return media.Author + " - " + media.Title
}

type suggestion struct {
	Artist string `json:"artiste"`
	Title  string `json:"titre"`
}

func parseSuggestions(answer string) []string {
	start := strings.Index(answer, "[")
	end := strings.LastIndex(answer, "]")
	if start < 0 || end < start {
		logger.Warn("Music: autoplay suggestions are not a JSON array")
		return nil
	}

	var items []suggestion
	if err := json.Unmarshal([]byte(answer[start:end+1]), &items); err != nil {
		logger.Warn("Music: autoplay parsing suggestions", "error", err)
		return nil
	}

	queries := make([]string, 0, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}
		if artist := strings.TrimSpace(item.Artist); artist != "" {
			title = artist + " - " + title
		}
		queries = append(queries, title)
		if len(queries) == autoplayPicks {
			break
		}
	}
	return queries
}

func resolveAll(ctx context.Context, queries []string, known map[string]bool) []audio.Media {
	found := make([]*audio.Media, len(queries))

	var wg sync.WaitGroup
	for i, query := range queries {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resolution, err := audio.Resolve(ctx, query)
			if err != nil || len(resolution.Media) == 0 {
				logger.Debug("Music: autoplay could not resolve suggestion", "query", query, "error", err)
				return
			}
			found[i] = &resolution.Media[0]
		}()
	}
	wg.Wait()

	picks := make([]audio.Media, 0, len(queries))
	for _, media := range found {
		if media == nil {
			continue
		}
		key := trackKey(*media)
		if known[key] {
			continue
		}
		known[key] = true
		picks = append(picks, *media)
	}
	return picks
}

func radioPicks(ctx context.Context, seed audio.Media, known map[string]bool) []audio.Media {
	related, err := audio.Radio(ctx, seed.URI)
	if err != nil {
		logger.Warn("Music: autoplay radio failed", "seed", seed.Title, "error", err)
		return nil
	}

	picks := make([]audio.Media, 0, autoplayPicks)
	for _, media := range related {
		key := trackKey(media)
		if known[key] {
			continue
		}
		known[key] = true
		picks = append(picks, media)
		if len(picks) == autoplayPicks {
			break
		}
	}
	return picks
}
