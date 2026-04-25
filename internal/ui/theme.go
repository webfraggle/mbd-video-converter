package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/gofont/gomonobolditalic"
	"golang.org/x/image/font/gofont/gomonoitalic"
	"golang.org/x/image/font/gofont/goregular"
)

// mbdTheme — "Departure Board".
//
// A dark industrial theme inspired by the warm amber of physical
// Zugzielanzeiger station displays. Deep neutrals carry the surface, a
// single accent (#E8A33D) marks state and primary action, type set on the
// Bigelow & Holmes Go fonts for a measured, technical feel.
type mbdTheme struct{}

func newTheme() fyne.Theme { return mbdTheme{} }

// Palette
var (
	cBackground  = color.NRGBA{0x14, 0x16, 0x1A, 0xFF}
	cInputBg     = color.NRGBA{0x1F, 0x21, 0x25, 0xFF}
	cButton      = color.NRGBA{0x26, 0x29, 0x2E, 0xFF}
	cHover       = color.NRGBA{0x33, 0x37, 0x3D, 0xFF}
	cPressed     = color.NRGBA{0x3F, 0x43, 0x48, 0xFF}
	cDisabledBtn = color.NRGBA{0x18, 0x1A, 0x1E, 0xFF}
	cBorder      = color.NRGBA{0x2A, 0x2D, 0x32, 0xFF}
	cFg          = color.NRGBA{0xF0, 0xED, 0xE5, 0xFF}
	cFgDisabled  = color.NRGBA{0x55, 0x58, 0x5C, 0xFF}
	cPlaceholder = color.NRGBA{0x6B, 0x6E, 0x73, 0xFF}
	cAmber       = color.NRGBA{0xE8, 0xA3, 0x3D, 0xFF}
	cAmberSoft   = color.NRGBA{0xE8, 0xA3, 0x3D, 0x33}
	cDanger      = color.NRGBA{0xE6, 0x6B, 0x6B, 0xFF}
	cShadow      = color.NRGBA{0x00, 0x00, 0x00, 0x40}
)

// Bundled fonts (Go font family by Bigelow & Holmes — already in our dependency tree).
var (
	fontRegular    = fyne.NewStaticResource("Go-Regular.ttf", goregular.TTF)
	fontBold       = fyne.NewStaticResource("Go-Bold.ttf", gobold.TTF)
	fontItalic     = fyne.NewStaticResource("Go-Italic.ttf", goitalic.TTF)
	fontBoldItalic = fyne.NewStaticResource("Go-BoldItalic.ttf", gobolditalic.TTF)
	fontMono       = fyne.NewStaticResource("Go-Mono.ttf", gomono.TTF)
	fontMonoBold   = fyne.NewStaticResource("Go-MonoBold.ttf", gomonobold.TTF)
	fontMonoItalic = fyne.NewStaticResource("Go-MonoItalic.ttf", gomonoitalic.TTF)
	fontMonoBI     = fyne.NewStaticResource("Go-MonoBoldItalic.ttf", gomonobolditalic.TTF)
)

func (mbdTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return cBackground
	case theme.ColorNameButton:
		return cButton
	case theme.ColorNameDisabled:
		return cFgDisabled
	case theme.ColorNameDisabledButton:
		return cDisabledBtn
	case theme.ColorNameError:
		return cDanger
	case theme.ColorNameFocus:
		return cAmber
	case theme.ColorNameForeground:
		return cFg
	case theme.ColorNameHover:
		return cHover
	case theme.ColorNameInputBackground:
		return cInputBg
	case theme.ColorNameInputBorder:
		return cBorder
	case theme.ColorNamePlaceHolder:
		return cPlaceholder
	case theme.ColorNamePressed:
		return cPressed
	case theme.ColorNamePrimary:
		return cAmber
	case theme.ColorNameScrollBar:
		return cButton
	case theme.ColorNameSelection:
		return cAmberSoft
	case theme.ColorNameSeparator:
		return cBorder
	case theme.ColorNameShadow:
		return cShadow
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (mbdTheme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Monospace {
		switch {
		case s.Bold && s.Italic:
			return fontMonoBI
		case s.Bold:
			return fontMonoBold
		case s.Italic:
			return fontMonoItalic
		default:
			return fontMono
		}
	}
	switch {
	case s.Bold && s.Italic:
		return fontBoldItalic
	case s.Bold:
		return fontBold
	case s.Italic:
		return fontItalic
	default:
		return fontRegular
	}
}

func (mbdTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (mbdTheme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 9
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameScrollBar:
		return 10
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 13
	case theme.SizeNameSubHeadingText:
		return 15
	case theme.SizeNameHeadingText:
		return 20
	case theme.SizeNameCaptionText:
		return 11
	}
	return theme.DefaultTheme().Size(n)
}
