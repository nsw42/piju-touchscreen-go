package screenblankmgr

type ProfileNone struct {
}

func (profile *ProfileNone) OnStartPlaying() {
	// Do nothing except implement the interface
}

func (profile *ProfileNone) OnStopPlaying() {
	// Do nothing except implement the interface
}

func (profile *ProfileNone) OnPlayingTick() {
	// Do nothing except implement the interface
}

func (profile *ProfileNone) OnStoppedDelayed() {
	// Do nothing except implement the interface
}
