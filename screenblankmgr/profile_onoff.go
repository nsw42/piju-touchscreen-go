package screenblankmgr

type ProfileOnOff struct {
}

func (profile *ProfileOnOff) OnStartPlaying() {
	setTimeout(60 * 60)
}

func (profile *ProfileOnOff) OnStopPlaying() {
	runXset("on")
	setTimeout(10)
}

func (profile *ProfileOnOff) OnPlayingTick() {
	runXset("off")
	runXset("reset")
}

func (profile *ProfileOnOff) OnStoppedDelayed() {
	blankScreenNow()
}
