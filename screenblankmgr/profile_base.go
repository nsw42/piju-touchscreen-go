package screenblankmgr

type ProfileBase interface {
	OnStartPlaying()
	OnStopPlaying()
	OnPlayingTick()
	OnStoppedDelayed()
}
