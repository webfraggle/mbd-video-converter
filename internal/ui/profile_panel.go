package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/profile"
)

// labeledEntry composes a small label on the left of a single-line entry.
// The label is wrapped in a fixed-width container so that across all rows of
// the profile form the entries start at the same x-position and end up the
// same width.
const profileLabelWidth = 92

func labeledEntry(label string, e *widget.Entry) fyne.CanvasObject {
	l := widget.NewLabel(label)
	return container.NewBorder(nil, nil, minSized(fyne.NewSize(profileLabelWidth, 0), l), nil, e)
}

type ProfilePanel struct {
	store    *profile.Store
	all      []profile.Profile
	selected int

	list                                                    *widget.List
	widthE, heightE, fpsE, qualityE, satE, gammaE, scalerE *widget.Entry
	advE                                                    *widget.Entry
	saveBtn, saveAsBtn, dupBtn, delBtn, newBtn              *widget.Button
	root                                                    *fyne.Container

	OnSelectionChanged func(p profile.Profile)
}

func NewProfilePanel(store *profile.Store) *ProfilePanel {
	pp := &ProfilePanel{store: store, selected: -1}
	all, _ := store.All()
	pp.all = all

	pp.list = widget.NewList(
		func() int { return len(pp.all) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("")
			lbl.Wrapping = fyne.TextWrapOff
			lbl.Truncation = fyne.TextTruncateEllipsis
			// Pad each row vertically so descenders (y, g, p) don't get
			// clipped against the row separator. theme.Padding contributes
			// ~6px top + 6px bottom of breathing room.
			return container.NewPadded(lbl)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			lbl := o.(*fyne.Container).Objects[0].(*widget.Label)
			p := pp.all[i]
			if p.Factory {
				lbl.SetText("○   " + p.Name)
				lbl.TextStyle = fyne.TextStyle{}
			} else {
				lbl.SetText("●   " + p.Name)
				lbl.TextStyle = fyne.TextStyle{Bold: true}
			}
		},
	)
	pp.list.OnSelected = func(i widget.ListItemID) {
		pp.selected = i
		pp.refreshFields()
		if pp.OnSelectionChanged != nil {
			pp.OnSelectionChanged(pp.all[i])
		}
	}

	pp.widthE = widget.NewEntry()
	pp.heightE = widget.NewEntry()
	pp.fpsE = widget.NewEntry()
	pp.qualityE = widget.NewEntry()
	pp.satE = widget.NewEntry()
	pp.gammaE = widget.NewEntry()
	pp.scalerE = widget.NewEntry()
	pp.advE = widget.NewMultiLineEntry()
	pp.advE.Wrapping = fyne.TextWrapWord
	pp.advE.SetMinRowsVisible(3)

	// Two-column compact form for the numeric fields.
	dimensionsRow := container.NewGridWithColumns(2,
		labeledEntry(i18n.T("profile.field.width"), pp.widthE),
		labeledEntry(i18n.T("profile.field.height"), pp.heightE),
	)
	rateQualityRow := container.NewGridWithColumns(2,
		labeledEntry(i18n.T("profile.field.fps"), pp.fpsE),
		labeledEntry(i18n.T("profile.field.quality"), pp.qualityE),
	)
	colorRow := container.NewGridWithColumns(2,
		labeledEntry(i18n.T("profile.field.saturation"), pp.satE),
		labeledEntry(i18n.T("profile.field.gamma"), pp.gammaE),
	)
	scalerRow := labeledEntry(i18n.T("profile.field.scaler"), pp.scalerE)

	form := container.NewVBox(
		container.NewPadded(subsectionLabel(i18n.T("profile.section.encoding"))),
		dimensionsRow,
		rateQualityRow,
		colorRow,
		scalerRow,
	)

	advAcc := widget.NewAccordion(
		widget.NewAccordionItem(i18n.T("profile.advanced"), pp.advE),
	)

	pp.newBtn = widget.NewButtonWithIcon(i18n.T("profile.btn.new"), theme.ContentAddIcon(), pp.onNew)
	pp.dupBtn = widget.NewButtonWithIcon(i18n.T("profile.btn.dup"), theme.ContentCopyIcon(), pp.onDup)
	pp.delBtn = widget.NewButtonWithIcon(i18n.T("profile.btn.del"), theme.DeleteIcon(), pp.onDel)
	pp.saveBtn = widget.NewButtonWithIcon(i18n.T("profile.btn.save"), theme.DocumentSaveIcon(), pp.onSave)
	pp.saveAsBtn = widget.NewButtonWithIcon(i18n.T("profile.btn.saveAs"), theme.DocumentCreateIcon(), pp.onSaveAs)
	pp.saveBtn.Importance = widget.HighImportance

	listButtons := container.NewGridWithColumns(3, pp.newBtn, pp.dupBtn, pp.delBtn)
	formButtons := container.NewGridWithColumns(2, pp.saveBtn, pp.saveAsBtn)

	header := container.NewVBox(
		container.NewPadded(sectionLabel(i18n.T("profile.section.title"))),
		amberLine(),
	)

	listSection := container.NewBorder(
		nil, container.NewPadded(listButtons),
		nil, nil,
		minSized(fyne.NewSize(0, 168), pp.list),
	)

	bottomSection := container.NewVBox(
		widget.NewSeparator(),
		container.NewPadded(form),
		container.NewPadded(advAcc),
		container.NewPadded(formButtons),
	)

	pp.root = container.NewBorder(
		header,
		bottomSection,
		nil, nil,
		listSection,
	)

	if len(pp.all) > 0 {
		pp.list.Select(0)
	}
	return pp
}

func (pp *ProfilePanel) Container() fyne.CanvasObject { return pp.root }

func (pp *ProfilePanel) refreshFields() {
	if pp.selected < 0 || pp.selected >= len(pp.all) {
		return
	}
	p := pp.all[pp.selected]
	pp.widthE.SetText(fmt.Sprintf("%d", p.Width))
	pp.heightE.SetText(fmt.Sprintf("%d", p.Height))
	pp.fpsE.SetText(fmt.Sprintf("%d", p.FPS))
	pp.qualityE.SetText(fmt.Sprintf("%d", p.Quality))
	pp.satE.SetText(fmt.Sprintf("%g", p.Saturation))
	pp.gammaE.SetText(fmt.Sprintf("%g", p.Gamma))
	pp.scalerE.SetText(p.Scaler)

	pp.saveBtn.Disable()
	if !p.Factory {
		pp.saveBtn.Enable()
	}
	pp.delBtn.Disable()
	if !p.Factory {
		pp.delBtn.Enable()
	}
}

// Selected returns the currently selected profile merged with editor edits.
func (pp *ProfilePanel) Selected() profile.Profile {
	if pp.selected < 0 || pp.selected >= len(pp.all) {
		return profile.Profile{}
	}
	p := pp.all[pp.selected]
	if v, err := atoi(pp.widthE.Text); err == nil {
		p.Width = v
	}
	if v, err := atoi(pp.heightE.Text); err == nil {
		p.Height = v
	}
	if v, err := atoi(pp.fpsE.Text); err == nil {
		p.FPS = v
	}
	if v, err := atoi(pp.qualityE.Text); err == nil {
		p.Quality = v
	}
	if v, err := atof(pp.satE.Text); err == nil {
		p.Saturation = v
	}
	if v, err := atof(pp.gammaE.Text); err == nil {
		p.Gamma = v
	}
	if pp.scalerE.Text != "" {
		p.Scaler = pp.scalerE.Text
	}
	return p
}

func (pp *ProfilePanel) AdvancedArgs() string { return pp.advE.Text }

func (pp *ProfilePanel) reload(selectID string) {
	all, _ := pp.store.All()
	pp.all = all
	pp.list.Refresh()
	for i, p := range pp.all {
		if p.ID == selectID {
			pp.list.Select(i)
			return
		}
	}
	if len(pp.all) > 0 {
		pp.list.Select(0)
	}
}

func (pp *ProfilePanel) currentUserList() []profile.Profile {
	out := []profile.Profile{}
	for _, p := range pp.all {
		if !p.Factory {
			out = append(out, p)
		}
	}
	return out
}

func (pp *ProfilePanel) saveUserList(users []profile.Profile, selectID string) {
	if err := pp.store.Save(users); err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	pp.reload(selectID)
}

func (pp *ProfilePanel) onNew() {
	pp.promptName("", func(name string) {
		newP := profile.Profile{
			ID: "user:" + uuid.NewString(), Name: name,
			Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos",
		}
		users := append(pp.currentUserList(), newP)
		pp.saveUserList(users, newP.ID)
	})
}

func (pp *ProfilePanel) onDup() {
	if pp.selected < 0 {
		return
	}
	src := pp.all[pp.selected]
	pp.promptName(src.Name+" (Kopie)", func(name string) {
		dup := src
		dup.Factory = false
		dup.ID = "user:" + uuid.NewString()
		dup.Name = name
		users := append(pp.currentUserList(), dup)
		pp.saveUserList(users, dup.ID)
	})
}

func (pp *ProfilePanel) onDel() {
	if pp.selected < 0 {
		return
	}
	cur := pp.all[pp.selected]
	if cur.Factory {
		return
	}
	users := pp.currentUserList()
	out := users[:0]
	for _, p := range users {
		if p.ID != cur.ID {
			out = append(out, p)
		}
	}
	pp.saveUserList(out, "")
}

func (pp *ProfilePanel) onSave() {
	if pp.selected < 0 {
		return
	}
	cur := pp.Selected()
	if cur.Factory {
		return
	}
	users := pp.currentUserList()
	for i := range users {
		if users[i].ID == cur.ID {
			users[i] = cur
		}
	}
	pp.saveUserList(users, cur.ID)
}

func (pp *ProfilePanel) onSaveAs() {
	pp.promptName("Mein Profil", func(name string) {
		base := pp.Selected()
		base.Factory = false
		base.ID = "user:" + uuid.NewString()
		base.Name = name
		users := append(pp.currentUserList(), base)
		pp.saveUserList(users, base.ID)
	})
}

func (pp *ProfilePanel) promptName(initial string, ok func(string)) {
	entry := widget.NewEntry()
	entry.SetText(initial)
	dialog.ShowForm("Profilname", "OK", "Abbrechen",
		[]*widget.FormItem{widget.NewFormItem("Name", entry)},
		func(confirm bool) {
			if !confirm || entry.Text == "" {
				return
			}
			ok(entry.Text)
		},
		fyne.CurrentApp().Driver().AllWindows()[0],
	)
}
