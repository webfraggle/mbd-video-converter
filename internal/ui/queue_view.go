package ui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/job"
)

type QueueView struct {
	jobs []*job.Job
	tbl  *widget.Table

	outDirEntry  *widget.Entry
	patternEntry *widget.Entry

	addBtn, clearBtn, cancelBtn, convertBtn *widget.Button
	root                                    *fyne.Container

	OnConvert   func()
	OnCancel    func()
	OnCancelJob func(jobID string)
}

func NewQueueView() *QueueView {
	qv := &QueueView{}

	qv.tbl = widget.NewTable(
		func() (int, int) { return len(qv.jobs) + 1, 4 }, // +1 for header
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("")
			lbl.Wrapping = fyne.TextWrapOff
			lbl.Truncation = fyne.TextTruncateEllipsis
			return lbl
		},
		func(id widget.TableCellID, c fyne.CanvasObject) {
			lbl := c.(*widget.Label)
			lbl.TextStyle = fyne.TextStyle{}
			if id.Row == 0 {
				// Header row — small caps caption look via bold style.
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				switch id.Col {
				case 0:
					lbl.SetText("№")
				case 1:
					lbl.SetText(i18n.T("queue.header.file"))
				case 2:
					lbl.SetText(i18n.T("queue.header.status"))
				case 3:
					lbl.SetText("")
				}
				return
			}
			j := qv.jobs[id.Row-1]
			switch id.Col {
			case 0:
				lbl.TextStyle = fyne.TextStyle{Monospace: true}
				lbl.SetText(fmt.Sprintf("%02d", id.Row))
			case 1:
				lbl.TextStyle = fyne.TextStyle{Monospace: true}
				lbl.SetText(filepath.Base(j.InputPath))
			case 2:
				lbl.SetText(statusGlyph(j))
				if j.Status == job.StatusFailed {
					lbl.TextStyle = fyne.TextStyle{Bold: true}
				}
			case 3:
				lbl.Alignment = fyne.TextAlignCenter
				lbl.SetText("✕")
			}
		},
	)
	qv.tbl.SetColumnWidth(0, 44)
	qv.tbl.SetColumnWidth(1, 320)
	qv.tbl.SetColumnWidth(2, 130)
	qv.tbl.SetColumnWidth(3, 36)
	qv.tbl.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 && id.Col == 3 && qv.OnCancelJob != nil {
			qv.OnCancelJob(qv.jobs[id.Row-1].ID)
		}
		qv.tbl.UnselectAll()
	}

	qv.addBtn = widget.NewButtonWithIcon(i18n.T("queue.btn.add"), theme.ContentAddIcon(), qv.onAdd)
	qv.clearBtn = widget.NewButtonWithIcon(i18n.T("queue.btn.clear"), theme.ContentClearIcon(), func() {
		qv.jobs = nil
		qv.tbl.Refresh()
	})
	qv.cancelBtn = widget.NewButtonWithIcon(i18n.T("queue.btn.cancel"), theme.CancelIcon(), func() {
		if qv.OnCancel != nil {
			qv.OnCancel()
		}
	})
	qv.convertBtn = widget.NewButtonWithIcon(i18n.T("queue.btn.convert"), theme.MediaPlayIcon(), func() {
		if qv.OnConvert != nil {
			qv.OnConvert()
		}
	})
	qv.convertBtn.Importance = widget.HighImportance

	leftButtons := container.NewHBox(qv.addBtn, qv.clearBtn)
	rightButtons := container.NewHBox(qv.cancelBtn, qv.convertBtn)
	bottomBar := container.NewBorder(nil, nil, leftButtons, rightButtons)

	qv.outDirEntry = widget.NewEntry()
	qv.outDirEntry.SetPlaceHolder(i18n.T("output.placeholder"))
	qv.patternEntry = widget.NewEntry()
	qv.patternEntry.SetText("{name}_{profile}_{fps}fps")

	pickBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), qv.onPickOutDir)
	outDirRow := container.NewBorder(
		nil, nil,
		widget.NewLabel(i18n.T("output.label.dir")),
		pickBtn,
		qv.outDirEntry,
	)
	patternRow := container.NewBorder(
		nil, nil,
		widget.NewLabel(i18n.T("output.label.pattern")),
		nil,
		qv.patternEntry,
	)

	header := container.NewVBox(
		container.NewPadded(sectionLabel(i18n.T("queue.section.title"))),
		amberLine(),
	)

	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewPadded(bottomBar),
		container.NewPadded(outDirRow),
		container.NewPadded(patternRow),
	)

	qv.root = container.NewBorder(header, footer, nil, nil, qv.tbl)
	return qv
}

func (qv *QueueView) Container() fyne.CanvasObject { return qv.root }
func (qv *QueueView) AddJob(j *job.Job)             { qv.jobs = append(qv.jobs, j); qv.tbl.Refresh() }
func (qv *QueueView) Jobs() []*job.Job              { return qv.jobs }
func (qv *QueueView) OutputDir() string             { return qv.outDirEntry.Text }
func (qv *QueueView) FilenamePattern() string       { return qv.patternEntry.Text }
func (qv *QueueView) Refresh()                      { qv.tbl.Refresh() }

// RemoveJob drops the job with the given id from the queue view and refreshes.
// No-op if the id is not in the current jobs slice.
func (qv *QueueView) RemoveJob(id string) {
	for i, j := range qv.jobs {
		if j.ID == id {
			qv.jobs = append(qv.jobs[:i], qv.jobs[i+1:]...)
			qv.tbl.Refresh()
			return
		}
	}
}

// statusGlyph returns the user-facing status string with a leading symbol so
// states are scannable without color: ▸ running, ✓ done, ✕ failed, ⌧ cancelled.
func statusGlyph(j *job.Job) string {
	switch j.Status {
	case job.StatusPending:
		return "·  " + i18n.T("queue.status.pending")
	case job.StatusRunning:
		if j.Progress > 0 {
			return fmt.Sprintf("▸  %.0f%%", j.Progress*100)
		}
		return "▸  " + i18n.T("queue.status.running")
	case job.StatusDone:
		return "✓  " + i18n.T("queue.status.done")
	case job.StatusFailed:
		return "✕  " + i18n.T("queue.status.failed")
	case job.StatusCancelled:
		return "⌧  " + i18n.T("queue.status.cancelled")
	}
	return ""
}

func (qv *QueueView) onPickOutDir() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		qv.outDirEntry.SetText(uri.Path())
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}

func (qv *QueueView) onAdd() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		qv.AddJob(job.NewJob(reader.URI().Path()))
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}
