package music

import "Eve/internal/bot/router"

func Register(r *router.Router) {
	if !Enabled() {
		return
	}

	r.OnCommand(CommandPlay, HandlePlay)
	r.OnCommand(CommandQueue, HandleQueue)
	r.OnCommand(CommandNowPlaying, HandleNowPlaying)
	r.OnCommand(CommandSkip, HandleSkip)
	r.OnCommand(CommandBack, HandleBack)
	r.OnCommand(CommandPause, HandlePause)
	r.OnCommand(CommandResume, HandleResume)
	r.OnCommand(CommandStop, HandleStop)
	r.OnCommand(CommandLoop, HandleLoop)
	r.OnCommand(CommandClear, HandleClear)
	r.OnCommand(CommandRemove, HandleRemove)
	r.OnCommand(CommandFilter, HandleFilter)
	r.OnCommand(CommandSyncedLyrics, HandleSyncedLyrics)

	r.OnButton(CustomIDBack, HandleBackButton)
	r.OnButton(CustomIDSkip, HandleSkipButton)
	r.OnButton(CustomIDPlayPause, HandlePlayPauseButton)
	r.OnButton(CustomIDLoop, HandleLoopButton)
	r.OnButton(CustomIDAutoplay, HandleAutoplayButton)
}
