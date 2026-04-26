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
	ffmpegPath.SetPlaceHolder(i18n.T("settings.ffmpegPath.placeholder"))

	defaultOutDir := widget.NewEntry()
	defaultOutDir.SetText(current.DefaultOutDir)

	pattern := widget.NewEntry()
	pattern.SetText(stringDefault(current.FilenamePattern, "{name}_{profile}_{fps}fps"))

	// OnExist: keep canonical internal values ("overwrite"|"suffix"|"fail")
	// in settings.json, but show localized labels in the Select.
	onExistInternal := []string{"overwrite", "suffix", "fail"}
	onExistLabels := []string{
		i18n.T("settings.onExist.overwrite"),
		i18n.T("settings.onExist.suffix"),
		i18n.T("settings.onExist.fail"),
	}
	labelToInternal := map[string]string{}
	internalToLabel := map[string]string{}
	for i, internal := range onExistInternal {
		labelToInternal[onExistLabels[i]] = internal
		internalToLabel[internal] = onExistLabels[i]
	}
	onExist := widget.NewSelect(onExistLabels, nil)
	onExist.SetSelected(internalToLabel[stringDefault(current.OnExist, "overwrite")])

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
		{Text: i18n.T("settings.label.language"), Widget: lang},
		{Text: i18n.T("settings.label.ffmpegPath"), Widget: ffmpegPath},
		{Text: i18n.T("settings.label.defaultOut"), Widget: defaultOutDir},
		{Text: i18n.T("settings.label.pattern"), Widget: pattern},
		{Text: i18n.T("settings.label.onExist"), Widget: onExist},
		{Text: "", Widget: openLogBtn},
		{Text: "", Widget: versionLabel},
	}

	d := dialog.NewForm(
		i18n.T("settings.title"),
		i18n.T("dialog.btn.ok"),
		i18n.T("dialog.btn.cancel"),
		form,
		func(ok bool) {
			if !ok {
				return
			}
			updated := settings.Settings{
				Language:        lang.Selected,
				FFmpegPath:      ffmpegPath.Text,
				DefaultOutDir:   defaultOutDir.Text,
				FilenamePattern: pattern.Text,
				OnExist:         labelToInternal[onExist.Selected],
				LastProfileID:   current.LastProfileID,
			}
			if err := store.Save(updated); err != nil {
				dialog.ShowError(err, parent)
				return
			}
			i18n.SetLanguage(updated.Language)
			onSaved(updated)
		},
		parent,
	)
	d.Resize(fyne.NewSize(560, 420))
	d.Show()
}
