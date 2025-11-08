package main

import (
	"fmt"
	"os"
	"strings"

	"net/http"
	_ "net/http/pprof"

	"github.com/akamensky/argparse"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"nsw42/piju-touchscreen-go/apiclient"
	"nsw42/piju-touchscreen-go/mainwindow"
	"nsw42/piju-touchscreen-go/screenblankmgr"
)

type Arguments struct {
	Debug bool
	Host  string
	PProf bool
	// Options related to the main window
	DarkMode           bool
	FullScreen         bool
	FixedLayout        bool
	CloseButton        bool
	HideMousePointer   bool
	ScreenBlankProfile screenblankmgr.ProfileBase
}

var args Arguments
var mainWindow *mainwindow.MainWindow
var apiClient *apiclient.Client
var screenMgr *screenblankmgr.ScreenBlankManager

func parseArgs() bool {
	var defaultHost string
	if hostname, err := os.Hostname(); err == nil {
		defaultHost = hostname + ":5000"
	}
	parser := argparse.NewParser("piju-touchscreen", "A GTK-based touchscreen UI for piju")
	debugArg := parser.Flag("", "debug", &argparse.Options{Default: false, Help: "Enable debug output"})
	hostArg := parser.String("", "host", &argparse.Options{Default: defaultHost, Help: "Connect to server at the given address"})
	pprofArg := parser.Flag("", "pprof", &argparse.Options{Default: false, Help: "Enable profiling server on port 6060"})
	modeArg := parser.Selector("m", "mode", []string{"dark", "light"}, &argparse.Options{Default: "light", Help: "Select the colour scheme of the UI: dark or light"})
	fullscreenArg := parser.Flag("", "fullscreen", &argparse.Options{Default: false, Help: "Show the main window full-screen"})
	layoutArg := parser.Selector("l", "layout", []string{"dynamic", "fixed"}, &argparse.Options{Default: "dynamic", Help: "Select whether to use a fixed layout or a dynamic layout to position controls"})
	closeButtonArg := parser.Flag("", "closebutton", &argparse.Options{Default: false, Help: "Show a close button"})
	hideMouseArg := parser.Flag("", "hidemousepointer", &argparse.Options{Default: false, Help: "Hide the mouse pointer when it is in the window"})
	screenblankArg := parser.Selector("", "screenblanker-profile", []string{"none", "balanced", "onoff"}, &argparse.Options{Default: "none", Help: "Actively manage the screen blank based on playpack state"})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Println(err)
		fmt.Print(parser.Usage(err))
		return false
	}

	args.Debug = *debugArg
	args.Host = *hostArg
	args.PProf = *pprofArg
	args.DarkMode = (*modeArg == "dark")
	args.FullScreen = *fullscreenArg
	args.FixedLayout = (*layoutArg == "fixed")
	args.CloseButton = *closeButtonArg
	args.HideMousePointer = *hideMouseArg
	switch *screenblankArg {
	case "none":
		args.ScreenBlankProfile = &screenblankmgr.ProfileNone{}
	case "balanced":
		args.ScreenBlankProfile = &screenblankmgr.ProfileBalanced{}
	case "onoff":
		args.ScreenBlankProfile = &screenblankmgr.ProfileOnOff{}
	}

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

	if args.PProf {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

	apiClient = &apiclient.Client{IsConnected: false, Host: args.Host}
	screenMgr = screenblankmgr.NewScreenBlankManager(args.ScreenBlankProfile)

	app := gtk.NewApplication("com.github.nsw42.piju-touchscreen-go", gio.ApplicationFlagsNone)
	app.ConnectActivate(func() { activate(app) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *gtk.Application) {
	mainWindow = mainwindow.NewMainWindow(app,
		apiClient,
		args.DarkMode,
		args.FullScreen,
		args.FixedLayout,
		args.CloseButton,
		args.HideMousePointer)

	glib.TimeoutAdd(5000, func() bool {
		if !apiClient.IsConnected {
			mainWindow.ShowNowPlaying(apiclient.NowPlaying{Status: apiclient.Error})
			apiClient.ConnectWS(mainWindow.QueueShowNowPlaying)
		}
		return glib.SOURCE_CONTINUE // =please keep calling me
	})
	glib.TimeoutAdd(1000, func() bool {
		mainWindow.CheckWindowSize()
		screenMgr.SetState(apiClient.PlayerStatus)
		return glib.SOURCE_CONTINUE // =please keep calling me
	})
}
