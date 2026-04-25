package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/job"
)

var acceptedExt = map[string]struct{}{
	".mp4": {}, ".mov": {}, ".avi": {}, ".mkv": {}, ".webm": {}, ".m4v": {},
}

type DropOverlay struct {
	bg    *canvas.Rectangle
	label *widget.Label
	root  *fyne.Container
}

func NewDropOverlay() *DropOverlay {
	bg := canvas.NewRectangle(color.NRGBA{R: 59, G: 110, B: 165, A: 80})
	lbl := widget.NewLabelWithStyle("⬇ "+i18n.T("dnd.overlay"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	d := &DropOverlay{
		bg:    bg,
		label: lbl,
		root:  container.NewStack(bg, container.NewCenter(lbl)),
	}
	d.root.Hide()
	return d
}

func (d *DropOverlay) Show()                          { d.root.Show() }
func (d *DropOverlay) Hide()                          { d.root.Hide() }
func (d *DropOverlay) CanvasObject() fyne.CanvasObject { return d.root }

// FilterAccepted picks only video URIs and returns the local paths.
func FilterAccepted(uris []fyne.URI) []string {
	out := make([]string, 0, len(uris))
	for _, u := range uris {
		if u == nil {
			continue
		}
		ext := strings.ToLower(u.Extension())
		if _, ok := acceptedExt[ext]; ok {
			out = append(out, u.Path())
		}
	}
	return out
}

// JobsFromPaths builds Job objects from a list of paths (used by both Add-File and Drop paths).
func JobsFromPaths(paths []string) []*job.Job {
	jobs := make([]*job.Job, 0, len(paths))
	for _, p := range paths {
		jobs = append(jobs, job.NewJob(p))
	}
	return jobs
}
