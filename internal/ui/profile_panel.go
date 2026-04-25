package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/webfraggle/mbd-video-converter/internal/i18n"
	"github.com/webfraggle/mbd-video-converter/internal/profile"
)

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
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			lbl := o.(*widget.Label)
			p := pp.all[i]
			prefix := ""
			if p.Factory {
				prefix = "🔒 "
			}
			lbl.SetText(prefix + p.Name)
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

	form := widget.NewForm(
		widget.NewFormItem(i18n.T("profile.field.width"), pp.widthE),
		widget.NewFormItem(i18n.T("profile.field.height"), pp.heightE),
		widget.NewFormItem(i18n.T("profile.field.fps"), pp.fpsE),
		widget.NewFormItem(i18n.T("profile.field.quality"), pp.qualityE),
		widget.NewFormItem(i18n.T("profile.field.saturation"), pp.satE),
		widget.NewFormItem(i18n.T("profile.field.gamma"), pp.gammaE),
		widget.NewFormItem(i18n.T("profile.field.scaler"), pp.scalerE),
	)

	advAcc := widget.NewAccordion(
		widget.NewAccordionItem(i18n.T("profile.advanced"), pp.advE),
	)

	pp.newBtn = widget.NewButton(i18n.T("profile.btn.new"), pp.onNew)
	pp.dupBtn = widget.NewButton(i18n.T("profile.btn.dup"), pp.onDup)
	pp.delBtn = widget.NewButton(i18n.T("profile.btn.del"), pp.onDel)
	pp.saveBtn = widget.NewButton(i18n.T("profile.btn.save"), pp.onSave)
	pp.saveAsBtn = widget.NewButton(i18n.T("profile.btn.saveAs"), pp.onSaveAs)

	listButtons := container.NewGridWithColumns(3, pp.newBtn, pp.dupBtn, pp.delBtn)
	formButtons := container.NewGridWithColumns(2, pp.saveBtn, pp.saveAsBtn)

	pp.root = container.NewBorder(
		container.NewVBox(widget.NewLabel(i18n.T("profile.header")), pp.list, listButtons),
		formButtons,
		nil, nil,
		container.NewVBox(form, advAcc),
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

// Stub handlers — wired up in Task 17.
func (pp *ProfilePanel) onNew()    {}
func (pp *ProfilePanel) onDup()    {}
func (pp *ProfilePanel) onDel()    {}
func (pp *ProfilePanel) onSave()   {}
func (pp *ProfilePanel) onSaveAs() {}
