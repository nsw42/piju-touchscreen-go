package apiclient

type NowPlaying struct {
	Status      Status // If this is Error, no other values in the struct can be relied upon
	IsTrack     bool
	ArtistName  string
	TrackName   string
	StreamName  string
	TrackNumber int
	AlbumTracks int
	ArtworkUri  string
	Artwork     []byte

	// TODO Volume   int
	// TODO Scanning string
}
