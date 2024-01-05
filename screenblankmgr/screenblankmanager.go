package screenblankmgr

import "nsw42/piju-touchscreen-go/apiclient"

const (
	tickInterval     = 5
	delayStopTimeout = 10
)

type ScreenBlankManager struct {
	State         apiclient.Status
	Profile       ProfileBase
	TickCountdown int
}

func NewScreenBlankManager(profile ProfileBase) *ScreenBlankManager {
	return &ScreenBlankManager{
		apiclient.Error,
		profile,
		tickInterval,
	}
}

func (manager *ScreenBlankManager) SetState(newState apiclient.Status) {
	if (manager.State == apiclient.Playing && newState == apiclient.Playing) || (manager.State != apiclient.Playing && newState != apiclient.Playing) {
		// State is, to all intents and purposes, unchanged
		manager.TickCountdown -= 1
		if manager.TickCountdown == 0 {
			if manager.State == apiclient.Playing {
				// Playing: tick the profile
				manager.Profile.OnPlayingTick()
				manager.TickCountdown = tickInterval
			} else {
				// Not playing - notify profile of (delayed) stop
				manager.Profile.OnStoppedDelayed()
				// Don't reset the countdown: we only call it once
			}
		}
	} else {
		// State has changed
		manager.State = newState
		if newState == apiclient.Playing {
			manager.Profile.OnStartPlaying()
			manager.TickCountdown = tickInterval
		} else {
			manager.Profile.OnStopPlaying()
			manager.TickCountdown = delayStopTimeout
		}
	}
}
