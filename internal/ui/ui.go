package ui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/job"
	"github.com/webfraggle/mbd-video-converter/internal/profile"
	"github.com/webfraggle/mbd-video-converter/internal/version"
)

func Run() {
	i18n.SetLanguage(i18n.DefaultLanguage())

	cfgDir, _ := os.UserConfigDir()
	appCfgDir := filepath.Join(cfgDir, "MBD-Videoconverter")
	profileStore := profile.NewStore(filepath.Join(appCfgDir, "profiles.json"))
	profilePanel := NewProfilePanel(profileStore)

	a := app.NewWithID("de.modellbahn-displays.mbd-videoconverter")
	w := a.NewWindow(i18n.T("app.title") + " " + version.Version)
	w.Resize(fyne.NewSize(1000, 640))

	qv := NewQueueView()
	left := qv.Container()
	right := profilePanel.Container()

	split := container.NewHSplit(left, right)
	split.SetOffset(0.62)

	settingsBtn := widget.NewButton(i18n.T("settings.title")+"…", func() {
		// task 21
	})
	header := container.NewBorder(nil, nil, nil, settingsBtn, widget.NewLabel(i18n.T("app.title")))

	overlay := NewDropOverlay()
	stack := container.NewStack(
		container.NewBorder(header, nil, nil, nil, split),
		overlay.CanvasObject(),
	)
	w.SetContent(stack)

	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		defer overlay.Hide()
		for _, p := range FilterAccepted(uris) {
			qv.AddJob(job.NewJob(p))
		}
	})

	w.ShowAndRun()
}
