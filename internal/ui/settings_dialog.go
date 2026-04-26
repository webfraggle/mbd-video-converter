package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/settings"
	"github.com/webfraggle/mbd-video-converter/internal/version"
)

// currentLogFolder holds the directory containing debug.log; set by Run() at startup.
var currentLogFolder = ""

func SetLogFolder(p string) { currentLogFolder = p }
func logFolderPath() string { return currentLogFolder }

func ShowSettingsDialog(parent fyne.Window, store *settings.Store, current settings.Settings, onSaved func(settings.Settings)) {
	lang := widget.NewSelect([]string{"de", "en"}, nil)
	lang.SetSelected(stringDefault(current.Language, i18n.DefaultLanguage()))

	ffmpegPath := widget.NewEntry()
	ffmpegPath.SetText(current.FFmpegPath)
	ffmpegPath.SetPlaceHolder("(leer = beigelegtes ffmpeg)")

	defaultOutDir := widget.NewEntry()
	defaultOutDir.SetText(current.DefaultOutDir)

	pattern := widget.NewEntry()
	pattern.SetText(stringDefault(current.FilenamePattern, "{name}_{profile}_{fps}fps"))

	onExist := widget.NewSelect([]string{"overwrite", "suffix", "fail"}, nil)
	onExist.SetSelected(stringDefault(current.OnExist, "overwrite"))

	openLogBtn := widget.NewButton(i18n.T("settings.btn.openLog"), func() {
		path := logFolderPath()
		if path == "" {
			return
		}
		if err := openInFileManager(path); err != nil {
			// Fall back: hand the user the path to paste themselves.
			fyne.CurrentApp().Clipboard().SetContent(path)
			dialog.ShowInformation(i18n.T("settings.btn.openLog"), path, parent)
		}
	})
	versionLabel := widget.NewLabel("Version: " + version.Version)

	form := []*widget.FormItem{
		{Text: "Sprache", Widget: lang},
		{Text: "ffmpeg-Pfad", Widget: ffmpegPath},
		{Text: "Default Output", Widget: defaultOutDir},
		{Text: "Pattern", Widget: pattern},
		{Text: "Bei vorhandener Datei", Widget: onExist},
		{Text: "", Widget: openLogBtn},
		{Text: "", Widget: versionLabel},
	}

	d := dialog.NewForm(i18n.T("settings.title"), "OK", "Abbrechen", form, func(ok bool) {
		if !ok {
			return
		}
		updated := settings.Settings{
			Language:        lang.Selected,
			FFmpegPath:      ffmpegPath.Text,
			DefaultOutDir:   defaultOutDir.Text,
			FilenamePattern: pattern.Text,
			OnExist:         onExist.Selected,
			LastProfileID:   current.LastProfileID,
		}
		if err := store.Save(updated); err != nil {
			dialog.ShowError(err, parent)
			return
		}
		i18n.SetLanguage(updated.Language)
		onSaved(updated)
	}, parent)
	d.Resize(fyne.NewSize(560, 420))
	d.Show()
}
