package audio

import "github.com/disgoorg/disgolink/v4/lavalink"

type Filter struct {
	Name  string
	Label string
	set   func(*lavalink.Filters)
	clear func(*lavalink.Filters)
	on    func(lavalink.Filters) bool
}

var bassBoostBands = lavalink.Equalizer{0.25, 0.25, 0.20, 0.15, 0.10}

var filters = []Filter{
	{
		Name:  "bassboost",
		Label: "Bassboost",
		set:   func(f *lavalink.Filters) { bands := bassBoostBands; f.Equalizer = &bands },
		clear: func(f *lavalink.Filters) { f.Equalizer = nil },
		on:    func(f lavalink.Filters) bool { return f.Equalizer != nil },
	},
	{
		Name:  "nightcore",
		Label: "Nightcore",
		set: func(f *lavalink.Filters) {
			f.Timescale = &lavalink.Timescale{Speed: 1.2, Pitch: 1.2, Rate: 1}
		},
		clear: func(f *lavalink.Filters) { f.Timescale = nil },
		on:    func(f lavalink.Filters) bool { return f.Timescale != nil && f.Timescale.Pitch > 1 },
	},
	{
		Name:  "vaporwave",
		Label: "Vaporwave",
		set: func(f *lavalink.Filters) {
			f.Timescale = &lavalink.Timescale{Speed: 0.85, Pitch: 0.8, Rate: 1}
		},
		clear: func(f *lavalink.Filters) { f.Timescale = nil },
		on:    func(f lavalink.Filters) bool { return f.Timescale != nil && f.Timescale.Pitch < 1 },
	},
	{
		Name:  "karaoke",
		Label: "Karaoké",
		set: func(f *lavalink.Filters) {
			f.Karaoke = &lavalink.Karaoke{Level: 1, MonoLevel: 1, FilterBand: 220, FilterWidth: 100}
		},
		clear: func(f *lavalink.Filters) { f.Karaoke = nil },
		on:    func(f lavalink.Filters) bool { return f.Karaoke != nil },
	},
	{
		Name:  "tremolo",
		Label: "Trémolo",
		set: func(f *lavalink.Filters) {
			f.Tremolo = &lavalink.Tremolo{Frequency: 2, Depth: 0.5}
		},
		clear: func(f *lavalink.Filters) { f.Tremolo = nil },
		on:    func(f lavalink.Filters) bool { return f.Tremolo != nil },
	},
	{
		Name:  "vibrato",
		Label: "Vibrato",
		set: func(f *lavalink.Filters) {
			f.Vibrato = &lavalink.Vibrato{Frequency: 2, Depth: 0.5}
		},
		clear: func(f *lavalink.Filters) { f.Vibrato = nil },
		on:    func(f lavalink.Filters) bool { return f.Vibrato != nil },
	},
	{
		Name:  "rotation",
		Label: "Rotation (8D)",
		set:   func(f *lavalink.Filters) { f.Rotation = &lavalink.Rotation{RotationHz: 1} },
		clear: func(f *lavalink.Filters) { f.Rotation = nil },
		on:    func(f lavalink.Filters) bool { return f.Rotation != nil },
	},
	{
		Name:  "lowpass",
		Label: "Passe-bas",
		set:   func(f *lavalink.Filters) { f.LowPass = &lavalink.LowPass{Smoothing: 20} },
		clear: func(f *lavalink.Filters) { f.LowPass = nil },
		on:    func(f lavalink.Filters) bool { return f.LowPass != nil },
	},
	{
		Name:  "distortion",
		Label: "Distorsion",
		set: func(f *lavalink.Filters) {
			f.Distortion = &lavalink.Distortion{SinScale: 1, CosScale: 1, TanScale: 1, Scale: 1}
		},
		clear: func(f *lavalink.Filters) { f.Distortion = nil },
		on:    func(f lavalink.Filters) bool { return f.Distortion != nil },
	},
}

func Filters() []Filter {
	return filters
}

func FilterLabel(name string) (string, bool) {
	for _, f := range filters {
		if f.Name == name {
			return f.Label, true
		}
	}
	return "", false
}

func ToggleFilter(current lavalink.Filters, name string) (lavalink.Filters, bool, bool) {
	for _, f := range filters {
		if f.Name != name {
			continue
		}
		if f.on(current) {
			f.clear(&current)
			return current, false, true
		}
		f.set(&current)
		return current, true, true
	}
	return current, false, false
}
