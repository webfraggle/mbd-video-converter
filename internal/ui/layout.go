package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// minSizeLayout enforces a minimum size on its single child while still
// expanding when the parent gives it more room. Used to give widget.List a
// usable visible height (Fyne's List defaults to a single-row MinSize).
type minSizeLayout struct{ min fyne.Size }

func (m *minSizeLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return m.min
}

func (m *minSizeLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objs {
		o.Resize(size)
		o.Move(fyne.NewPos(0, 0))
	}
}

func minSized(min fyne.Size, child fyne.CanvasObject) *fyne.Container {
	return container.New(&minSizeLayout{min: min}, child)
}

// sectionLabel renders a small uppercase amber heading with letter-spacing
// inspired by departure-board panel labels. RichText is used because Label
// has no per-instance text size or color.
func sectionLabel(text string) *widget.RichText {
	rt := widget.NewRichText(&widget.TextSegment{
		Text: " " + text + " ",
		Style: widget.RichTextStyle{
			ColorName: theme.ColorNamePrimary,
			SizeName:  theme.SizeNameCaptionText,
			TextStyle: fyne.TextStyle{Bold: true},
		},
	})
	rt.Wrapping = fyne.TextWrapOff
	return rt
}

// amberLine returns a thin horizontal rule using the amber accent —
// used as a section divider directly under sectionLabel headings.
func amberLine() *canvas.Rectangle {
	r := canvas.NewRectangle(cAmber)
	r.SetMinSize(fyne.NewSize(0, 1))
	return r
}
