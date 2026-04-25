package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, c fyne.CanvasObject) {
			lbl := c.(*widget.Label)
			if id.Row == 0 {
				switch id.Col {
				case 0:
					lbl.SetText("#")
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
				lbl.SetText(fmt.Sprintf("%d", id.Row))
			case 1:
				lbl.SetText(j.InputPath)
			case 2:
				lbl.SetText(statusLabel(j))
			case 3:
				lbl.SetText("✕")
			}
		},
	)
	qv.tbl.SetColumnWidth(0, 40)
	qv.tbl.SetColumnWidth(1, 360)
	qv.tbl.SetColumnWidth(2, 110)
	qv.tbl.SetColumnWidth(3, 30)
	qv.tbl.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 && id.Col == 3 && qv.OnCancelJob != nil {
			qv.OnCancelJob(qv.jobs[id.Row-1].ID)
		}
		qv.tbl.UnselectAll()
	}

	qv.addBtn = widget.NewButton(i18n.T("queue.btn.add"), qv.onAdd)
	qv.clearBtn = widget.NewButton(i18n.T("queue.btn.clear"), func() {
		qv.jobs = nil
		qv.tbl.Refresh()
	})
	qv.cancelBtn = widget.NewButton(i18n.T("queue.btn.cancel"), func() {
		if qv.OnCancel != nil {
			qv.OnCancel()
		}
	})
	qv.convertBtn = widget.NewButton(i18n.T("queue.btn.convert"), func() {
		if qv.OnConvert != nil {
			qv.OnConvert()
		}
	})

	topButtons := container.NewHBox(qv.addBtn, qv.clearBtn)
	rightButtons := container.NewHBox(qv.cancelBtn, qv.convertBtn)
	bottomBar := container.NewBorder(nil, nil, topButtons, rightButtons)

	qv.outDirEntry = widget.NewEntry()
	qv.outDirEntry.SetPlaceHolder("(leer = neben Input)")
	qv.patternEntry = widget.NewEntry()
	qv.patternEntry.SetText("{name}_{profile}_{fps}fps")

	outDirRow := container.NewBorder(nil, nil, widget.NewLabel(i18n.T("output.label.dir")), widget.NewButton("…", qv.onPickOutDir), qv.outDirEntry)
	patternRow := container.NewBorder(nil, nil, widget.NewLabel(i18n.T("output.label.pattern")), nil, qv.patternEntry)

	qv.root = container.NewBorder(
		nil,
		container.NewVBox(bottomBar, outDirRow, patternRow),
		nil, nil,
		qv.tbl,
	)
	return qv
}

func (qv *QueueView) Container() fyne.CanvasObject { return qv.root }
func (qv *QueueView) AddJob(j *job.Job)             { qv.jobs = append(qv.jobs, j); qv.tbl.Refresh() }
func (qv *QueueView) Jobs() []*job.Job              { return qv.jobs }
func (qv *QueueView) OutputDir() string             { return qv.outDirEntry.Text }
func (qv *QueueView) FilenamePattern() string       { return qv.patternEntry.Text }
func (qv *QueueView) Refresh()                      { qv.tbl.Refresh() }

func statusLabel(j *job.Job) string {
	switch j.Status {
	case job.StatusPending:
		return i18n.T("queue.status.pending")
	case job.StatusRunning:
		if j.Progress > 0 {
			return fmt.Sprintf("%.0f%%", j.Progress*100)
		}
		return i18n.T("queue.status.running")
	case job.StatusDone:
		return i18n.T("queue.status.done")
	case job.StatusFailed:
		return i18n.T("queue.status.failed")
	case job.StatusCancelled:
		return i18n.T("queue.status.cancelled")
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
