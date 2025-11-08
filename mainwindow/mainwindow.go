package mainwindow

import (
	"embed"
	"log"
	"net/url"
	"nsw42/piju-touchscreen-go/apiclient"
	"slices"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	qrcode "github.com/skip2/go-qrcode"
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

type MainWindowState int

const (
	MainWindowStateControls MainWindowState = iota
	MainWindowStateQRCode
)

type MainWindow struct {
	State             MainWindowState
	ControlsContainer *gtk.Widget
	QrCodeContainer   *gtk.Fixed
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
	MenuButton        *gtk.MenuButton
	CloseButton       *gtk.Button
	ScanningIndicator *gtk.Image
	PauseIcon         *gtk.Image
	PlayIcon          *gtk.Image
	PrevIcon          *gtk.Image
	NextIcon          *gtk.Image
	QrCodeIcon        *gtk.Image
	// MenuIcon          *gtk.Image
	MenuAction        *gio.SimpleAction
	PreviousWidth     int
	PreviousHeight    int
	HideMousePointer  bool
	PlayPauseAction   func()
	CurrentArtworkUri string
}

//go:embed icons/*.png
var icons embed.FS

//go:embed mainwindow.css
var cssString string

func findThemeIcon(widget *gtk.Widget, iconNames []string) string {
	display := widget.Display()
	theme := gtk.IconThemeGetForDisplay(display)
	for _, iconName := range iconNames {
		if slices.Contains(theme.IconNames(), iconName) {
			return iconName
		}
	}
	return iconNames[0]
}

func imageFromEmbedPNG(leafName string) *gtk.Image {
	iconData, err := icons.ReadFile("icons/" + leafName)
	if err != nil {
		log.Fatalf("icons.ReadFile: %w", err)
	}

	return imageFromPNGBytes(iconData)
}

func imageFromPNGBytes(data []byte) *gtk.Image {
	l, err := gdkpixbuf.NewPixbufLoaderWithType("png")
	if err != nil {
		log.Fatalf("NewLoaderWithType png: %w", err)
	}
	defer l.Close()

	if err := l.Write(data); err != nil {
		log.Fatalf("PixbufLoader.Write: %w", err)
	}

	pixbuf := l.Pixbuf()
	return gtk.NewImageFromPixbuf(pixbuf)
}

func loadLocalImageNoMode(iconName string, iconSize int) *gtk.Image {
	leafName := iconName
	if iconSize != 0 {
		leafName += "_" + strconv.Itoa(iconSize)
	}
	leafName += ".png"
	return imageFromEmbedPNG(leafName)
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
	return imageFromEmbedPNG(leafName)
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

func (window *MainWindow) layoutFixed() {
	fixedContainer := gtk.NewFixed()
	var xPadding, y0Padding, labelH float64
	xPadding = 10
	y0Padding = 10
	labelH = maxImageSize / 2

	controlsContainer := gtk.NewFixed()
	window.ControlsContainer = &controlsContainer.Widget

	controlsContainer.Put(window.Artwork, xPadding, y0Padding)

	trackArtistX0 := xPadding + maxImageSize + xPadding
	controlsContainer.Put(window.TrackNameLabel, trackArtistX0, y0Padding)

	artistY0 := y0Padding + labelH + y0Padding
	controlsContainer.Put(window.ArtistLabel, trackArtistX0, artistY0)

	for _, label := range []*gtk.Label{window.TrackNameLabel, window.ArtistLabel} {
		label.SetSizeRequest(int(screenWidth-trackArtistX0-xPadding), int(labelH))
	}

	controlsContainer.Put(window.NoTrackLabel, (screenWidth-noTrackLabelW)/2, 150)
	window.NoTrackLabel.SetSizeRequest(int(noTrackLabelW), 32)
	// buttons
	// image is 100x100; button padding takes it to 112x110
	// (on macOS, at least)
	//   SPC  IMG  2xSPC  IMG  2xSPC  IMG  SPC
	// 6xSPC + 3xIMG = SCREEN_WIDTH
	// => SPC = (SCREEN_WIDTH - 3*IMG) / 6
	controlsContainer.Put(window.PrevButton, buttonXPadding, buttonY0)
	controlsContainer.Put(window.PlayPauseButton, (screenWidth-imgButtonW)/2, buttonY0)
	controlsContainer.Put(window.NextButton, screenWidth-buttonXPadding-imgButtonW, buttonY0)

	fixedContainer.Put(controlsContainer, 0, 0)

	fixedContainer.Put(window.QrCodeContainer, 0, 0)

	fixedContainer.Put(window.ScanningIndicator, screenWidth-20, 4)

	fixedContainer.Put(window.MenuButton, 0, 0)

	if window.CloseButton != nil {
		fixedContainer.Put(window.CloseButton, 0, 0)
	}

	window.Window.SetChild(fixedContainer)
}

func (window *MainWindow) layoutDynamic() {
	margin := 20

	window.Artwork.SetMarginStart(margin)
	window.Artwork.SetMarginEnd(margin)
	window.Artwork.SetMarginTop(margin)
	window.Artwork.SetMarginBottom(margin)

	for _, button := range []*gtk.Button{window.PrevButton, window.PlayPauseButton, window.NextButton} {
		button.SetMarginStart(margin)
		button.SetMarginEnd(margin)
		button.SetMarginTop(margin)
		button.SetMarginBottom(margin)
	}

	trackArtistContainer := gtk.NewBox(gtk.OrientationVertical, margin)
	trackArtistContainer.Append(window.TrackNameLabel)
	trackArtistContainer.Append(window.ArtistLabel)
	trackArtistContainer.SetVExpand(true)

	trackArtistContainer.SetMarginStart(margin)
	trackArtistContainer.SetMarginEnd(margin)
	trackArtistContainer.SetMarginTop(margin)
	trackArtistContainer.SetMarginBottom(margin)

	topRowContainer := gtk.NewBox(gtk.OrientationHorizontal, margin)
	topRowContainer.Append(window.Artwork)
	topRowContainer.Append(trackArtistContainer)
	topRowContainer.SetVAlign(gtk.AlignCenter)
	topRowContainer.SetVExpand(true)

	bottomRowContainer := gtk.NewBox(gtk.OrientationHorizontal, margin)
	bottomRowContainer.Append(window.PrevButton)
	bottomRowContainer.Append(window.PlayPauseButton)
	bottomRowContainer.Append(window.NextButton)
	bottomRowContainer.SetVAlign(gtk.AlignStart)
	bottomRowContainer.SetHExpand(true)
	bottomRowContainer.SetHomogeneous(true)

	controlsContainer := gtk.NewBox(gtk.OrientationVertical, margin)
	controlsContainer.Append(topRowContainer)
	controlsContainer.Append(bottomRowContainer)
	controlsContainer.SetHomogeneous(false)
	window.ControlsContainer = &controlsContainer.Widget

	overlay := gtk.NewOverlay()
	overlay.AddOverlay(window.QrCodeContainer) // ensure it's the bottom in the z-order
	window.ScanningIndicator.SetHAlign(gtk.AlignEnd)
	window.ScanningIndicator.SetVAlign(gtk.AlignStart)
	window.ScanningIndicator.SetMarginEnd(margin)
	window.ScanningIndicator.SetMarginTop(margin)
	overlay.AddOverlay(window.ScanningIndicator)
	window.MenuButton.SetHAlign(gtk.AlignStart)
	window.MenuButton.SetVAlign(gtk.AlignStart)
	window.MenuButton.SetMarginStart(margin / 2)
	window.MenuButton.SetMarginTop(margin / 2)
	overlay.AddOverlay(window.MenuButton)
	if window.CloseButton != nil {
		window.CloseButton.SetHAlign(gtk.AlignStart)
		window.CloseButton.SetVAlign(gtk.AlignStart)
		overlay.AddOverlay(window.CloseButton)
	}
	overlay.SetChild(controlsContainer)

	window.Window.SetChild(overlay)
}

func NewMainWindow(app *gtk.Application,
	apiClient *apiclient.Client,
	darkMode bool,
	fullScreen bool,
	fixedLayout bool,
	closeButton bool,
	hideMousePointer bool,
) *MainWindow {

	rtn := &MainWindow{}
	rtn.State = MainWindowStateControls
	rtn.ApiClient = apiClient
	rtn.DarkMode = darkMode
	rtn.HideMousePointer = hideMousePointer

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
	cssProvider := gtk.NewCSSProvider()
	cssProvider.LoadFromString(cssString)
	gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

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

	// Menu button
	rtn.MenuButton = gtk.NewMenuButton()
	rtn.MenuButton.SetHAlign(gtk.AlignCenter)
	menu := gio.NewMenu()
	menu.Append("Local music", "app.resume('local')")
	menu.Append("Radio", "app.resume('radio')")
	menu.Append("Link", "app.resume('link')")
	rtn.MenuButton.SetMenuModel(menu)
	rtn.MenuButton.Popover().SetHasArrow(false)
	if darkMode {
		rtn.MenuButton.AddCSSClass("piju-dark-button")
	}
	rtn.MenuAction = gio.NewSimpleActionStateful("resume", glib.NewVariantType("s"), glib.NewVariantString("local"))
	app.ActionMap.AddAction(rtn.MenuAction)
	rtn.MenuAction.ConnectActivate(func(param *glib.Variant) {
		resumeType := param.String()
		if resumeType == "link" {
			// touchscreen-only action: show the link QR code
			rtn.State = MainWindowStateQRCode
			rtn.ControlsContainer.SetVisible(false)
			rtn.QrCodeContainer.SetVisible(true)
			rtn.MenuAction.ChangeState(glib.NewVariantString("link"))
		} else {
			rtn.State = MainWindowStateControls
			rtn.ControlsContainer.SetVisible(true)
			rtn.QrCodeContainer.SetVisible(false)
			rtn.ApiClient.SendResumeType(resumeType)
		}
	})

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

	// Overlays
	rtn.ScanningIndicator = loadLocalImageNoMode("circle", 16)
	if closeButton {
		closeIcon := loadLocalImageNoMode("window-close", 0)
		rtn.CloseButton = gtk.NewButton()
		rtn.CloseButton.SetChild(closeIcon)
		rtn.CloseButton.ConnectClicked(rtn.OnQuit)
	}

	// Layout and show
	rtn.QrCodeContainer = gtk.NewFixed()
	rtn.QrCodeContainer.SetVisible(false)
	if fixedLayout {
		rtn.NoTrackLabel = mkLabel(gtk.JustifyCenter, false, darkMode)
		rtn.layoutFixed()
	} else {
		rtn.NoTrackLabel = rtn.ArtistLabel
		rtn.layoutDynamic()
	}

	window.ConnectRealize(rtn.OnRealized)
	window.SetVisible(true)

	rtn.ShowNowPlaying(apiclient.NowPlaying{Status: apiclient.Stopped})
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

func (window *MainWindow) OnQuit() {
	window.Window.Destroy()
}

func (window *MainWindow) OnRealized() {
	if window.HideMousePointer {
		window.Window.SetCursor(gdk.NewCursorFromName("none", nil))
	}

	var iconSize int
	if window.Window.AllocatedWidth() > 1000 {
		iconSize = 200
	} else {
		iconSize = 100
	}

	// SetChild not present in the gtk bindings, even though it's in the GTK docs
	// window.MenuIcon = loadLocalImage("bars", window.DarkMode, iconSize)
	// window.MenuButton.SetChild(window.MenuIcon)

	menuIconName := findThemeIcon(&window.MenuButton.Widget, []string{"view-more-horizontal-symbolic", "open-menu-symbolic", "xfce-em-menu"})
	window.MenuButton.SetIconName(menuIconName)

	window.PauseIcon = loadLocalImage("pause", window.DarkMode, iconSize)
	window.PauseIcon.SetParent(window.PlayPauseButton)

	window.PlayIcon = loadLocalImage("play", window.DarkMode, iconSize)
	window.PlayIcon.SetParent(window.PlayPauseButton)

	window.PrevIcon = loadLocalImage("backward", window.DarkMode, iconSize)
	window.PrevIcon.SetParent(window.PrevButton)

	window.NextIcon = loadLocalImage("forward", window.DarkMode, iconSize)
	window.NextIcon.SetParent(window.NextButton)
}

func (window *MainWindow) Resized(newWidth, newHeight int) {
	qrSize := min(newWidth, newHeight) - 32
	qrSize = max(qrSize, 16)
	if window.QrCodeIcon != nil {
		window.QrCodeContainer.Remove(window.QrCodeIcon)
	}
	window.QrCodeIcon = window.NewQRCode(qrSize)
	window.QrCodeIcon.SetVisible(true)
	window.QrCodeIcon.SetSizeRequest(qrSize, qrSize)
	x := (newWidth - qrSize) / 2
	y := (newHeight - qrSize) / 2
	window.QrCodeContainer.Put(window.QrCodeIcon, float64(x), float64(y))
	window.PreviousWidth = newWidth
	window.PreviousHeight = newHeight
}

func (window *MainWindow) NewQRCode(size int) *gtk.Image {
	webui, _ := url.Parse(window.ApiClient.Host)
	webui.Host = webui.Hostname() // strip the port
	webuiUrl := webui.String()
	png, err := qrcode.Encode(webuiUrl, qrcode.Medium, size)
	if err != nil {
		log.Println("Error generating QR code:", err)
	}
	return imageFromPNGBytes(png)
}

func (window *MainWindow) showConnectionError() {
	window.ArtistLabel.SetVisible(false)
	window.TrackNameLabel.SetVisible(false)
	window.Artwork.SetVisible(false)
	window.NoTrackLabel.SetVisible(true)
	window.NoTrackLabel.SetLabel("Connection error")
	window.ScanningIndicator.SetVisible(false)
	window.PlayIcon.SetVisible(true)
	window.PauseIcon.SetVisible(false)
	window.PrevButton.SetSensitive(false)
	window.PlayPauseButton.SetSensitive(false)
	window.NextButton.SetSensitive(false)
}

func (window *MainWindow) QueueShowNowPlaying(nowPlaying apiclient.NowPlaying) {
	glib.IdleAdd(func() bool {
		window.ShowNowPlaying(nowPlaying)
		return glib.SOURCE_REMOVE // =no need to call me again
	})
}

func (window *MainWindow) ShowNowPlaying(nowPlaying apiclient.NowPlaying) {
	if nowPlaying.Status == apiclient.Error {
		window.showConnectionError()
	} else {
		window.showNowPlayingArtistAndTrack(nowPlaying)
		window.showNowPlayingImage(nowPlaying)
		window.showNowPlayingPlayPauseIcon(nowPlaying)
		window.showNowPlayingPrevNext(nowPlaying)
		window.showNowPlayingLocalRadio(nowPlaying)
		window.ScanningIndicator.SetVisible(nowPlaying.Scanning)
	}
}

func (window *MainWindow) CheckWindowSize() {
	newWidth := window.Window.Width()
	newHeight := window.Window.Height()
	if newWidth != window.PreviousWidth || newHeight != window.PreviousHeight {
		window.Resized(newWidth, newHeight)
	}
}

func (window *MainWindow) showNowPlayingArtistAndTrack(nowPlaying apiclient.NowPlaying) {
	if nowPlaying.IsTrack {
		window.NoTrackLabel.SetVisible(false)
		window.ArtistLabel.SetLabel(nowPlaying.ArtistName)
		window.ArtistLabel.SetVisible(true)
		window.TrackNameLabel.SetLabel(nowPlaying.TrackName)
		window.TrackNameLabel.SetVisible(true)
	} else if nowPlaying.StreamName != "" {
		window.NoTrackLabel.SetVisible(false)
		window.ArtistLabel.SetVisible(false)
		window.TrackNameLabel.SetLabel(nowPlaying.StreamName)
		window.TrackNameLabel.SetVisible(true)
	} else {
		window.ArtistLabel.SetVisible(false)
		window.TrackNameLabel.SetVisible(false)
		window.NoTrackLabel.SetLabel("No track")
		window.NoTrackLabel.SetVisible(true)
	}
}

func (window *MainWindow) showNowPlayingImage(nowPlaying apiclient.NowPlaying) {
	if nowPlaying.ArtworkUri == window.CurrentArtworkUri {
		// Ensure the artwork is visible; otherwise, there is nothing to do
		if nowPlaying.Artwork != nil {
			window.Artwork.SetVisible(true)
		}
		return
	}
	if !window.showNowPlayingImageInner(nowPlaying) {
		// Either no artwork or it's corrupted
		window.Artwork.SetVisible(false)
	}
	window.CurrentArtworkUri = nowPlaying.ArtworkUri
}

func (window *MainWindow) showNowPlayingImageInner(nowPlaying apiclient.NowPlaying) bool {
	// Returns true if there is a valid image, false otherwise
	if nowPlaying.Artwork == nil {
		return false
	}

	loader := gdkpixbuf.NewPixbufLoader()
	if loader == nil {
		log.Println("Failed to allocate pixbuf loader")
		return false
	}
	if err := loader.Write(nowPlaying.Artwork); err != nil {
		log.Println("loader.Write failed:", err.Error())
		return false
	}
	if err := loader.Close(); err != nil {
		log.Println("loader.Close failed:", err.Error())
		return false
	}
	pixbuf := loader.Pixbuf()
	if pixbuf == nil {
		log.Println("loader.Pixbuf failed")
		return false
	}

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
	window.Artwork.SetVisible(true)
	return true
}

func (window *MainWindow) showNowPlayingLocalRadio(nowPlaying apiclient.NowPlaying) {
	var state string
	if nowPlaying.StreamName == "" {
		state = "local"
	} else {
		state = "radio"
	}
	if !window.MenuButton.Popover().IsVisible() && window.State == MainWindowStateControls {
		window.MenuAction.ChangeState(glib.NewVariantString(state))
	}
}

func (window *MainWindow) showNowPlayingPlayPauseIcon(nowPlaying apiclient.NowPlaying) {
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

func (window *MainWindow) showNowPlayingPrevNext(nowPlaying apiclient.NowPlaying) {
	window.PrevButton.SetSensitive(nowPlaying.TrackNumber > 1)
	window.NextButton.SetSensitive(nowPlaying.TrackNumber > 0 &&
		nowPlaying.AlbumTracks > 0 &&
		nowPlaying.TrackNumber < nowPlaying.AlbumTracks)
}
