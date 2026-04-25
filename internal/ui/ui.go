package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/ffmpeg"
	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/job"
	"github.com/webfraggle/mbd-video-converter/internal/logx"
	"github.com/webfraggle/mbd-video-converter/internal/profile"
	"github.com/webfraggle/mbd-video-converter/internal/settings"
	"github.com/webfraggle/mbd-video-converter/internal/version"
)

func Run() {
	i18n.SetLanguage(i18n.DefaultLanguage())

	cfgDir, _ := os.UserConfigDir()
	appCfgDir := filepath.Join(cfgDir, "MBD-Videoconverter")

	profileStore := profile.NewStore(filepath.Join(appCfgDir, "profiles.json"))
	profilePanel := NewProfilePanel(profileStore)

	settingsStore := settings.New(filepath.Join(appCfgDir, "settings.json"))
	appSettings, _ := settingsStore.Load()
	i18n.SetLanguage(stringDefault(appSettings.Language, i18n.DefaultLanguage()))

	logPath, _ := logx.Setup(appCfgDir)
	SetLogFolder(filepath.Dir(logPath))
	log.Printf("MBD-Videoconverter %s starting (log: %s)", version.Version, logPath)

	bundleDir := executableDir()

	a := app.NewWithID("de.modellbahn-displays.mbd-videoconverter")
	a.Settings().SetTheme(newTheme())
	w := a.NewWindow(i18n.T("app.title") + " " + version.Version)
	w.Resize(fyne.NewSize(1040, 680))

	qv := NewQueueView()
	left := qv.Container()
	right := profilePanel.Container()

	split := container.NewHSplit(left, right)
	split.SetOffset(0.62)

	var cancelHandle atomic.Value

	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		ShowSettingsDialog(w, settingsStore, appSettings, func(updated settings.Settings) {
			appSettings = updated
			dialog.ShowInformation(i18n.T("settings.restartHint.title"), i18n.T("settings.restartHint.body"), w)
		})
	})

	header := buildHeader(settingsBtn)

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

	qv.OnConvert = func() {
		bin, err := ffmpeg.Locate(appSettings.FFmpegPath, bundleDir)
		if err != nil {
			dialog.ShowError(fmt.Errorf("%s", i18n.T("err.ffmpegNotFound")), w)
			return
		}
		sel := profilePanel.Selected()
		cfg := job.EncodingConfig{
			ProfileNameForPattern: profileNameForPattern(sel),
			Width:                 sel.Width,
			Height:                sel.Height,
			FPS:                   sel.FPS,
			Quality:               sel.Quality,
			Saturation:            sel.Saturation,
			Gamma:                 sel.Gamma,
			Scaler:                sel.Scaler,
			AdvancedArgs:          profilePanel.AdvancedArgs(),
			OutputDir:             qv.OutputDir(),
			FilenamePattern:       qv.FilenamePattern(),
			OnExist:               appSettings.OnExist,
		}
		q := job.NewQueue(&job.FFmpegRunner{BinaryPath: bin})
		for _, j := range qv.Jobs() {
			q.Add(j)
		}
		events := make(chan job.QueueEvent, 64)
		q.SubscribeEvents(events)

		ctx, cancel := context.WithCancel(context.Background())
		cancelHandle.Store(cancel)

		go q.Run(ctx, cfg)
		go func() {
			for range events {
				fyne.Do(qv.Refresh)
			}
		}()
	}

	qv.OnCancel = func() {
		if c, ok := cancelHandle.Load().(context.CancelFunc); ok && c != nil {
			c()
		}
	}

	qv.OnCancelJob = func(id string) {
		for _, j := range qv.Jobs() {
			if j.ID == id {
				j.Cancel()
				return
			}
		}
	}

	w.ShowAndRun()
}

// buildHeader returns the top app bar — wordmark with version, an amber
// hairline separator, and the settings icon at the trailing edge. The two
// rules are stacked: amber on top of a neutral separator for a stamped feel.
func buildHeader(settingsBtn fyne.CanvasObject) fyne.CanvasObject {
	wordmark := widget.NewRichText(&widget.TextSegment{
		Text: "MBD-VIDEOCONVERTER",
		Style: widget.RichTextStyle{
			SizeName:  theme.SizeNameSubHeadingText,
			TextStyle: fyne.TextStyle{Bold: true},
		},
	})
	wordmark.Wrapping = fyne.TextWrapOff

	versionTag := widget.NewRichText(&widget.TextSegment{
		Text: "  " + version.Version,
		Style: widget.RichTextStyle{
			ColorName: theme.ColorNamePrimary,
			SizeName:  theme.SizeNameCaptionText,
			TextStyle: fyne.TextStyle{Bold: true, Monospace: true},
		},
	})
	versionTag.Wrapping = fyne.TextWrapOff

	titleRow := container.NewBorder(
		nil, nil,
		container.NewHBox(wordmark, versionTag),
		settingsBtn,
		nil,
	)

	hair := canvas.NewRectangle(cAmber)
	hair.SetMinSize(fyne.NewSize(0, 1))

	return container.NewVBox(
		container.NewPadded(titleRow),
		hair,
	)
}
