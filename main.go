package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/version"
)

func main() {
	a := app.NewWithID("de.modellbahn-displays.mbd-videoconverter")
	w := a.NewWindow("MBD-Videoconverter " + version.Version)
	w.SetContent(widget.NewLabel("Hello, world."))
	w.Resize(fyne.NewSize(900, 600))
	w.ShowAndRun()
}
