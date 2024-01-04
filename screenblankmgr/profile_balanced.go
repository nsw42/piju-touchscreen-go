package screenblankmgr

type ProfileBalanced struct {
}

func (profile *ProfileBalanced) OnStartPlaying() {
	setTimeout(300)
}

func (profile *ProfileBalanced) OnStopPlaying() {
	setTimeout(30)
}

func (profile *ProfileBalanced) OnPlayingTick() {
	// Do nothing except implement the interface
}

func (profile *ProfileBalanced) OnStoppedDelayed() {
	blankScreenNow()
}
