package mainwindow

import (
	"nsw42/piju-touchscreen-go/apiclient"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	screenWidth  = 800
	screenHeight = 480

	maxImageSize = 300

	// Constants related to a fixed layout:
	noTrackLabelW  float64 = 200
	imgButtonW     float64 = 112
	imgButtonH     float64 = 110
	y1_padding     float64 = 20
	buttonY0       float64 = screenHeight - y1_padding - imgButtonH
	buttonXPadding float64 = (screenWidth - 3*imgButtonW) / 6
)

type MainWindow struct {
	ApiClient         *apiclient.Client
	DarkMode          bool
	Window            *gtk.ApplicationWindow
	Artwork           *gtk.Image
	TrackNameLabel    *gtk.Label
	NoTrackLabel      *gtk.Label
	ArtistLabel       *gtk.Label
	PrevButton        *gtk.Button
	PlayPauseButton   *gtk.Button
	NextButton        *gtk.Button
	PauseIcon         *gtk.Image
	PlayIcon          *gtk.Image
	PrevIcon          *gtk.Image
	NextIcon          *gtk.Image
	HideMousePointer  bool
	PlayPauseAction   func()
	CurrentArtworkUri string
}

func loadLocalImage(iconName string, darkMode bool, iconSize int) *gtk.Image {
	leafName := iconName
	if darkMode {
		leafName += "-dark"
	} else {
		leafName += "-light"
	}
	if iconSize != 0 {
		leafName += "_" + strconv.Itoa(iconSize)
	}
	leafName += ".png"
	// TODO: Absolute path?
	filename := "icons/" + leafName
	return gtk.NewImageFromFile(filename)
}

func mkLabel(justification gtk.Justification, large bool, darkMode bool) *gtk.Label {
	label := gtk.NewLabel("")
	label.SetHExpand(true)
	label.SetVExpand(true)
	label.SetWrap(true)
	var xalign float32
	if justification == gtk.JustifyLeft {
		xalign = 0.0
	} else if justification == gtk.JustifyRight {
		xalign = 1.0
	} else {
		xalign = 0.5
	}
	label.SetXAlign(xalign)

	var mode, size, class string
	if darkMode {
		mode = "dark"
	} else {
		mode = "light"
	}
	if large {
		size = "large"
	} else {
		size = "normal"
	}
	class = "piju-" + mode + "-" + size + "-label"
	label.AddCSSClass(class)

	return label
}

func layoutFixed(window *MainWindow) {
	fixedContainer := gtk.NewFixed()
	var xPadding, y0Padding, labelH float64
	xPadding = 10
	y0Padding = 10
	labelH = maxImageSize / 2

	fixedContainer.Put(window.Artwork, xPadding, y0Padding)

	trackArtistX0 := xPadding + maxImageSize + xPadding
	fixedContainer.Put(window.TrackNameLabel, trackArtistX0, y0Padding)

	artistY0 := y0Padding + labelH + y0Padding
	fixedContainer.Put(window.ArtistLabel, trackArtistX0, artistY0)

	for _, label := range []*gtk.Label{window.TrackNameLabel, window.ArtistLabel} {
		label.SetSizeRequest(int(screenWidth-trackArtistX0-xPadding), int(labelH))
	}

	fixedContainer.Put(window.NoTrackLabel, (screenWidth-noTrackLabelW)/2, 150)
	window.NoTrackLabel.SetSizeRequest(int(noTrackLabelW), 32)
	// buttons
	// image is 100x100; button padding takes it to 112x110
	// (on macOS, at least)
	//   SPC  IMG  2xSPC  IMG  2xSPC  IMG  SPC
	// 6xSPC + 3xIMG = SCREEN_WIDTH
	// => SPC = (SCREEN_WIDTH - 3*IMG) / 6
	fixedContainer.Put(window.PrevButton, buttonXPadding, buttonY0)
	fixedContainer.Put(window.PlayPauseButton, (screenWidth-imgButtonW)/2, buttonY0)
	fixedContainer.Put(window.NextButton, screenWidth-buttonXPadding-imgButtonW, buttonY0)

	// TODO fixedContainer.Put(window.scanning_indicator_icon, SCREEN_WIDTH-20, 4)

	window.Window.SetChild(fixedContainer)
}

func layoutDynamic(window *MainWindow) {
	// TODO
}

func NewMainWindow(app *gtk.Application, apiClient *apiclient.Client, darkMode bool, fullScreen bool, fixedLayout bool) *MainWindow {

	rtn := &MainWindow{}
	rtn.ApiClient = apiClient
	rtn.DarkMode = darkMode

	// Initialise each bit of the window in turn,
	// saving the results in rtn

	// The window itself:
	window := gtk.NewApplicationWindow(app)
	window.SetTitle("PiJu")
	if darkMode {
		window.AddCSSClass("piju-dark-background")
	}
	if fullScreen {
		window.Fullscreen()
	} else {
		window.SetSizeRequest(screenWidth, screenHeight)
	}
	css_provider := gtk.NewCSSProvider()
	css_provider.LoadFromPath("mainwindow/mainwindow.css")
	gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), css_provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	rtn.Window = window

	// The artwork
	artwork := gtk.NewImage()
	artwork.SetHExpand(false)
	artwork.SetVExpand(false)
	artwork.SetSizeRequest(maxImageSize, maxImageSize)
	rtn.Artwork = artwork

	// The labels
	rtn.TrackNameLabel = mkLabel(gtk.JustifyCenter, true, darkMode)
	rtn.ArtistLabel = mkLabel(gtk.JustifyCenter, false, darkMode)

	// Previous button
	rtn.PrevButton = gtk.NewButton()
	rtn.PrevButton.SetHAlign(gtk.AlignStart)
	rtn.PrevButton.ConnectClicked(rtn.OnPrevious)

	// Play/pause button
	rtn.PlayPauseButton = gtk.NewButton()
	rtn.PlayPauseButton.SetHAlign(gtk.AlignCenter)
	rtn.PlayPauseButton.ConnectClicked(rtn.OnPlayPause)

	// Next button
	rtn.NextButton = gtk.NewButton()
	rtn.NextButton.SetHAlign(gtk.AlignEnd)
	rtn.NextButton.ConnectClicked(rtn.OnNext)

	// Common properties to the buttons
	buttons := []*gtk.Button{rtn.PrevButton, rtn.PlayPauseButton, rtn.NextButton}
	for _, button := range buttons {
		button.SetFocusOnClick(false)
		button.SetVAlign(gtk.AlignCenter)
		button.SetSizeRequest(100, 100)
		if darkMode {
			button.AddCSSClass("piju-dark-button")
		}
	}

	// TODO: Scanning indicator

	if fixedLayout {
		rtn.NoTrackLabel = mkLabel(gtk.JustifyCenter, false, darkMode)
		layoutFixed(rtn)
	} else {
		rtn.NoTrackLabel = rtn.ArtistLabel
		layoutDynamic(rtn)
	}

	window.ConnectRealize(rtn.OnRealized)
	window.Show()

	rtn.ShowConnectionError() // TODO
	return rtn
}

func (window *MainWindow) OnNext() {
	window.ApiClient.SendNext()
}

func (window *MainWindow) OnPlayPause() {
	if window.PlayPauseAction != nil {
		window.PlayPauseAction()
	}
}

func (window *MainWindow) OnPrevious() {
	window.ApiClient.SendPrevious()
}

func (window *MainWindow) OnRealized() {
	// TODO: Hide mouse pointer
	var iconSize int
	if window.Window.AllocatedWidth() > 1000 {
		iconSize = 200
	} else {
		iconSize = 100
	}

	window.PauseIcon = loadLocalImage("pause", window.DarkMode, iconSize)
	window.PauseIcon.SetParent(window.PlayPauseButton)

	window.PlayIcon = loadLocalImage("play", window.DarkMode, iconSize)
	window.PlayIcon.SetParent(window.PlayPauseButton)

	window.PrevIcon = loadLocalImage("backward", window.DarkMode, iconSize)
	window.PrevIcon.SetParent(window.PrevButton)

	window.NextIcon = loadLocalImage("forward", window.DarkMode, iconSize)
	window.NextIcon.SetParent(window.NextButton)
}

func (window *MainWindow) ShowConnectionError() {
	window.ArtistLabel.Hide()
	window.TrackNameLabel.Hide()
	window.Artwork.Hide()
	window.NoTrackLabel.Show()
	window.NoTrackLabel.SetLabel("Connection error")
	// window.scanning_indicator_icon.hide()
	window.PlayIcon.SetVisible(true)
	window.PauseIcon.SetVisible(false)
	window.PrevButton.SetSensitive(false)
	window.PlayPauseButton.SetSensitive(false)
	window.NextButton.SetSensitive(false)
}

func (window *MainWindow) ShowNowPlaying(nowPlaying apiclient.NowPlaying) {
	if nowPlaying.Status == apiclient.Error {
		window.ShowConnectionError()
	} else {
		window.ShowNowPlayingArtistAndTrack(nowPlaying)
		window.ShowNowPlayingImage(nowPlaying)
		window.ShowNowPlayingPlayPauseIcon(nowPlaying)
		window.ShowNowPlayingPrevNext(nowPlaying)
	}
}

func (window *MainWindow) ShowNowPlayingArtistAndTrack(nowPlaying apiclient.NowPlaying) {
	if nowPlaying.IsTrack {
		window.NoTrackLabel.Hide()
		window.ArtistLabel.SetLabel(nowPlaying.ArtistName)
		window.ArtistLabel.Show()
		window.TrackNameLabel.SetLabel(nowPlaying.TrackName)
		window.TrackNameLabel.Show()
	} else if nowPlaying.StreamName != "" {
		window.NoTrackLabel.Hide()
		window.ArtistLabel.Hide()
		window.TrackNameLabel.SetLabel(nowPlaying.StreamName)
		window.TrackNameLabel.Show()
	} else {
		window.ArtistLabel.Hide()
		window.TrackNameLabel.Hide()
		window.NoTrackLabel.SetLabel("No track")
		window.NoTrackLabel.Show()
	}
}

func (window *MainWindow) ShowNowPlayingImage(nowPlaying apiclient.NowPlaying) {
	if nowPlaying.ArtworkUri == window.CurrentArtworkUri {
		// Ensure the artwork is visible; otherwise, there is nothing to do
		if nowPlaying.Artwork != nil {
			window.Artwork.Show()
		}
		return
	}
	if !window.showNowPlayingImageInner(nowPlaying) {
		// Either no artwork or it's corrupted
		window.Artwork.Hide()
	}
	window.CurrentArtworkUri = nowPlaying.ArtworkUri
}

func (window *MainWindow) showNowPlayingImageInner(nowPlaying apiclient.NowPlaying) bool {
	// Returns true if there is a valid image, false otherwise
	if nowPlaying.Artwork == nil {
		return false
	}

	loader := gdkpixbuf.NewPixbufLoader()
	err := loader.Write(nowPlaying.Artwork)
	if err == nil {
		err = loader.Close()
	}
	if err != nil {
		return false
	}
	pixbuf := loader.Pixbuf()

	width := pixbuf.Width()
	height := pixbuf.Height()
	if (width > maxImageSize) || (height > maxImageSize) {
		var destWidth, destHeight int
		if width > height {
			destWidth = maxImageSize
			destHeight = height * destWidth / width
		} else {
			destHeight = maxImageSize
			destWidth = width * destHeight / height
		}
		pixbuf = pixbuf.ScaleSimple(destWidth, destHeight, gdkpixbuf.InterpBilinear)
	}
	window.Artwork.SetFromPixbuf(pixbuf)
	window.Artwork.Show()
	return true
}

func (window *MainWindow) ShowNowPlayingPlayPauseIcon(nowPlaying apiclient.NowPlaying) {
	var sensitive bool
	var icon *gtk.Image
	var action func()

	switch nowPlaying.Status {
	case apiclient.Stopped:
		sensitive = false
		icon = window.PlayIcon
		action = nil
	case apiclient.Playing:
		sensitive = true
		icon = window.PauseIcon
		action = window.ApiClient.SendPause
	case apiclient.Paused:
		sensitive = true
		icon = window.PlayIcon
		action = window.ApiClient.SendResume
	}
	if icon == nil {
		// We're not yet fully initialised
		return
	}
	icon.SetVisible(true)
	var otherIcon *gtk.Image
	if icon == window.PlayIcon {
		otherIcon = window.PauseIcon
	} else {
		otherIcon = window.PlayIcon
	}
	otherIcon.SetVisible(false)
	window.PlayPauseButton.SetSensitive(sensitive)
	window.PlayPauseAction = action
}

func (window *MainWindow) ShowNowPlayingPrevNext(nowPlaying apiclient.NowPlaying) {
	window.PrevButton.SetSensitive(nowPlaying.TrackNumber > 1)
	window.NextButton.SetSensitive(nowPlaying.TrackNumber > 0 &&
		nowPlaying.AlbumTracks > 0 &&
		nowPlaying.TrackNumber < nowPlaying.AlbumTracks)
}