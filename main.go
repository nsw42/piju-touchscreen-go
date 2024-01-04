package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/akamensky/argparse"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"nsw42/piju-touchscreen-go/apiclient"
	"nsw42/piju-touchscreen-go/mainwindow"
)

type Arguments struct {
	Debug bool
	Host  string
	// Options related to the main window
	DarkMode         bool
	FullScreen       bool
	FixedLayout      bool
	CloseButton      bool
	HideMousePointer bool
}

var args Arguments
var mainWindow *mainwindow.MainWindow
var apiClient *apiclient.Client

func parseArgs() bool {
	parser := argparse.NewParser("piju-touchscreen", "A GTK-based touchscreen UI for piju")
	debugArg := parser.Flag("", "debug", &argparse.Options{Default: false, Help: "Enable debug output"})
	hostArg := parser.String("", "host", &argparse.Options{Default: "localhost:5000", Help: "Connect to server at the given address"})
	modeArg := parser.Selector("m", "mode", []string{"dark", "light"}, &argparse.Options{Default: "light", Help: "Select the colour scheme of the UI: dark or light"})
	fullscreenArg := parser.Flag("", "fullscreen", &argparse.Options{Default: false, Help: "Show the main window full-screen. Default is as a desktop window."})
	layoutArg := parser.Selector("l", "layout", []string{"dynamic", "fixed"}, &argparse.Options{Default: "dynamic", Help: "Select whether to use a fixed layout or a dynamic layout to position controls."})
	closeButtonArg := parser.Flag("", "closebutton", &argparse.Options{Default: false, Help: "Show a close button. Default is to rely on window furniture."})
	hideMouseArg := parser.Flag("", "hidemousepointer", &argparse.Options{Default: false, Help: "Hide the mouse pointer when it is in the window. Default is not to."})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Println(err)
		fmt.Print(parser.Usage(err))
		return false
	}

	args.Debug = *debugArg
	args.Host = *hostArg
	args.DarkMode = (*modeArg == "dark")
	args.FullScreen = *fullscreenArg
	args.FixedLayout = (*layoutArg == "fixed")
	args.CloseButton = *closeButtonArg
	args.HideMousePointer = *hideMouseArg

	if !strings.HasPrefix(args.Host, "http") {
		args.Host = "http://" + args.Host
	}
	if !strings.Contains(args.Host[6:], ":") {
		args.Host += ":5000"
	}
	if !strings.HasSuffix(args.Host, "/") {
		args.Host += "/"
	}

	// Prevent GTK from parsing the arguments
	os.Args = []string{}

	return true
}

func main() {
	if !parseArgs() {
		return
	}

	apiClient = &apiclient.Client{Host: args.Host}

	app := gtk.NewApplication("com.github.diamondburned.gotk4-examples.gtk4.simple", gio.ApplicationFlagsNone)
	app.ConnectActivate(func() { activate(app) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *gtk.Application) {
	mainWindow = mainwindow.NewMainWindow(app, apiClient, true, false, true)

	glib.TimeoutAdd(1000, getNowPlaying)
}

func getNowPlaying() {
	status := apiClient.GetCurrentStatus()
	fmt.Println(status.Status, status.ArtistName, status.TrackName, status.StreamName, status.TrackNumber, "/", status.AlbumTracks, status.Artwork[:20])

	glib.IdleAdd(func() { mainWindow.ShowNowPlaying(status) })

	// And update again, in another second
	glib.TimeoutAdd(1000, getNowPlaying)
}
