package music

import (
	"context"
	"sync"
	"time"

	"Eve/internal/audio"

	"github.com/disgoorg/snowflake/v2"
)

type RepeatMode int

const (
	RepeatOff RepeatMode = iota
	RepeatTrack
	RepeatQueue
)

func (m RepeatMode) Label() string {
	switch m {
	case RepeatTrack:
		return "Musique"
	case RepeatQueue:
		return "File d'attente"
	default:
		return "Désactivée"
	}
}

func (m RepeatMode) Icon() string {
	switch m {
	case RepeatTrack:
		return "🔂"
	case RepeatQueue:
		return "🔁"
	default:
		return ""
	}
}

type guildState struct {
	mu             sync.Mutex
	tracks         []audio.Media
	history        []audio.Media
	playing        *audio.Media
	repeat         RepeatMode
	textChannelID  snowflake.ID
	voiceChannelID snowflake.ID
	nowPlayingID   snowflake.ID
	lyricsThreadID snowflake.ID
	lyricsCancel   context.CancelFunc
	leaveTimer     *time.Timer
}

const maxHistory = 25

var (
	statesMu sync.Mutex
	states   = make(map[snowflake.ID]*guildState)
)

func stateFor(guildID snowflake.ID) *guildState {
	statesMu.Lock()
	defer statesMu.Unlock()

	state, ok := states[guildID]
	if !ok {
		state = &guildState{}
		states[guildID] = state
	}
	return state
}

func existingState(guildID snowflake.ID) (*guildState, bool) {
	statesMu.Lock()
	defer statesMu.Unlock()

	state, ok := states[guildID]
	return state, ok
}

func dropState(guildID snowflake.ID) {
	statesMu.Lock()
	state, ok := states[guildID]
	delete(states, guildID)
	statesMu.Unlock()

	if ok {
		state.cancelLeave()
	}
}

func (s *guildState) enqueue(tracks ...audio.Media) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tracks = append(s.tracks, tracks...)
}

func (s *guildState) enqueueFront(track audio.Media) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tracks = append([]audio.Media{track}, s.tracks...)
}

func (s *guildState) pop() (audio.Media, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.tracks) == 0 {
		return audio.Media{}, false
	}
	track := s.tracks[0]
	s.tracks = s.tracks[1:]
	return track, true
}

func (s *guildState) list() []audio.Media {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]audio.Media, len(s.tracks))
	copy(out, s.tracks)
	return out
}

func (s *guildState) size() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.tracks)
}

func (s *guildState) clear() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := len(s.tracks)
	s.tracks = nil
	return removed
}

func (s *guildState) removeAt(index int) (audio.Media, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= len(s.tracks) {
		return audio.Media{}, false
	}
	track := s.tracks[index]
	s.tracks = append(s.tracks[:index], s.tracks[index+1:]...)
	return track, true
}

func (s *guildState) pushHistory(track audio.Media) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = append(s.history, track)
	if len(s.history) > maxHistory {
		s.history = s.history[len(s.history)-maxHistory:]
	}
}

func (s *guildState) popHistory() (audio.Media, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.history) == 0 {
		return audio.Media{}, false
	}
	track := s.history[len(s.history)-1]
	s.history = s.history[:len(s.history)-1]
	return track, true
}

func (s *guildState) setCurrent(media *audio.Media) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playing = media
}

func (s *guildState) current() *audio.Media {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.playing == nil {
		return nil
	}
	copied := *s.playing
	return &copied
}

func (s *guildState) setRepeat(mode RepeatMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repeat = mode
}

func (s *guildState) cycleRepeat() RepeatMode {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.repeat == RepeatQueue {
		s.repeat = RepeatOff
	} else {
		s.repeat++
	}
	return s.repeat
}

func (s *guildState) repeatMode() RepeatMode {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repeat
}

func (s *guildState) setChannels(textChannelID snowflake.ID, voiceChannelID snowflake.ID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.textChannelID = textChannelID
	s.voiceChannelID = voiceChannelID
}

func (s *guildState) channels() (snowflake.ID, snowflake.ID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.textChannelID, s.voiceChannelID
}

func (s *guildState) setNowPlaying(messageID snowflake.ID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nowPlayingID = messageID
}

func (s *guildState) nowPlaying() snowflake.ID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nowPlayingID
}

func (s *guildState) setLyrics(threadID snowflake.ID, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lyricsThreadID = threadID
	s.lyricsCancel = cancel
}

func (s *guildState) takeLyrics() (snowflake.ID, context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	threadID, cancel := s.lyricsThreadID, s.lyricsCancel
	s.lyricsThreadID = 0
	s.lyricsCancel = nil
	return threadID, cancel
}

func (s *guildState) lyricsThread() snowflake.ID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lyricsThreadID
}

func (s *guildState) scheduleLeave(after time.Duration, fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.leaveTimer != nil {
		s.leaveTimer.Stop()
	}
	s.leaveTimer = time.AfterFunc(after, fn)
}

func (s *guildState) cancelLeave() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.leaveTimer != nil {
		s.leaveTimer.Stop()
		s.leaveTimer = nil
	}
}
