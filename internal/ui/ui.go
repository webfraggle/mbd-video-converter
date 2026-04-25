package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/version"
)

func Run() {
	i18n.SetLanguage(i18n.DefaultLanguage())

	a := app.NewWithID("de.modellbahn-displays.mbd-videoconverter")
	w := a.NewWindow(i18n.T("app.title") + " " + version.Version)
	w.Resize(fyne.NewSize(1000, 640))

	// Placeholders; subsequent tasks fill these in.
	left := container.NewBorder(widget.NewLabel(i18n.T("queue.header.file")), nil, nil, nil, widget.NewLabel("(queue placeholder)"))
	right := container.NewBorder(widget.NewLabel(i18n.T("profile.header")), nil, nil, nil, widget.NewLabel("(profile panel placeholder)"))

	split := container.NewHSplit(left, right)
	split.SetOffset(0.62)

	settingsBtn := widget.NewButton(i18n.T("settings.title")+"…", func() {
		// task 21
	})
	header := container.NewBorder(nil, nil, nil, settingsBtn, widget.NewLabel(i18n.T("app.title")))

	w.SetContent(container.NewBorder(header, nil, nil, nil, split))
	w.ShowAndRun()
}
