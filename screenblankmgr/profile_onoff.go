package screenblankmgr

type ProfileOnOff struct {
}

func (profile *ProfileOnOff) OnStartPlaying() {
	setTimeout(60 * 60)
	profile.OnPlayingTick()
}

func (profile *ProfileOnOff) OnStopPlaying() {
	setTimeout(10)
	runXset("on")
}

func (profile *ProfileOnOff) OnPlayingTick() {
	runXset("off")
	runXset("reset")
}

func (profile *ProfileOnOff) OnStoppedDelayed() {
	blankScreenNow()
}
