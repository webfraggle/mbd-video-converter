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

// sectionLabel renders the primary CI-orange section heading: a small,
// uppercase, bold label that introduces a major panel area (e.g.
// "WARTESCHLANGE", "AKTIVES PROFIL").
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

// subsectionLabel uses the secondary CI teal — a quieter heading marker
// that sits inside a section to call out a sub-area (e.g. "ENCODING" inside
// the active profile panel).
func subsectionLabel(text string) *widget.RichText {
	rt := widget.NewRichText(&widget.TextSegment{
		Text: " " + text + " ",
		Style: widget.RichTextStyle{
			ColorName: theme.ColorNameSuccess, // bound to cTeal in mbdTheme
			SizeName:  theme.SizeNameCaptionText,
			TextStyle: fyne.TextStyle{Bold: true},
		},
	})
	rt.Wrapping = fyne.TextWrapOff
	return rt
}

// primaryLine returns a 1-px horizontal rule in CI orange — used as a
// hairline directly under sectionLabel headings.
func primaryLine() *canvas.Rectangle {
	r := canvas.NewRectangle(cOrange)
	r.SetMinSize(fyne.NewSize(0, 1))
	return r
}

// amberLine kept as an alias for primaryLine so existing callers continue
// to compile during the CI rename.
func amberLine() *canvas.Rectangle { return primaryLine() }
