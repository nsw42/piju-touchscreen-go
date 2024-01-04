package apiclient

type Status int64

const (
	Error Status = iota
	Stopped
	Playing
	Paused
)

func (s Status) String() string {
	switch s {
	case Error:
		return "error"
	case Stopped:
		return "stopped"
	case Playing:
		return "playing"
	case Paused:
		return "paused"
	}
	return "???"
}
