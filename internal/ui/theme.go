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

// mbdTheme — "MBD CI".
//
// Brand-aligned light theme: white surface, near-black body text, orange
// (#FD7014) as the primary accent for headlines and active state, teal
// (#037F8C) as the supporting accent for subsection labels and the "done"
// state. Type set on the Bigelow & Holmes Go fonts.
type mbdTheme struct{}

func newTheme() fyne.Theme { return mbdTheme{} }

// Palette — pure neutrals + two CI accents.
var (
	cBackground  = color.NRGBA{0xFF, 0xFF, 0xFF, 0xFF}
	cInputBg     = color.NRGBA{0xFF, 0xFF, 0xFF, 0xFF}
	cButton      = color.NRGBA{0xF4, 0xF4, 0xF4, 0xFF}
	// Hover and Pressed are drawn as overlays *on top of* a button's
	// existing fill (neutral grey or CI orange). They must therefore be
	// translucent dark tints — an opaque grey would obliterate the orange
	// of HighImportance buttons on hover.
	cHover   = color.NRGBA{0x00, 0x00, 0x00, 0x14} // ~8% black
	cPressed = color.NRGBA{0x00, 0x00, 0x00, 0x28} // ~16% black
	cDisabledBtn = color.NRGBA{0xF9, 0xF9, 0xF9, 0xFF}
	cBorder      = color.NRGBA{0xE5, 0xE5, 0xE5, 0xFF}
	cInputBorder = color.NRGBA{0xD4, 0xD4, 0xD4, 0xFF}
	cFg          = color.NRGBA{0x17, 0x17, 0x17, 0xFF} // near-black body text
	cFgDisabled  = color.NRGBA{0xB5, 0xB5, 0xB5, 0xFF}
	cPlaceholder = color.NRGBA{0x9A, 0x9A, 0x9A, 0xFF}
	cOrange      = color.NRGBA{0xFD, 0x70, 0x14, 0xFF} // CI primary
	cOrangeSoft  = color.NRGBA{0xFD, 0x70, 0x14, 0x24}
	cTeal        = color.NRGBA{0x03, 0x7F, 0x8C, 0xFF} // CI secondary
	cDanger      = color.NRGBA{0xC0, 0x39, 0x2B, 0xFF}
	cShadow      = color.NRGBA{0x00, 0x00, 0x00, 0x14}
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

// cOnPrimary is used for text rendered on top of the orange/teal fills
// (e.g. HighImportance Convert and Save buttons). Plain white.
var cOnPrimary = color.NRGBA{0xFF, 0xFF, 0xFF, 0xFF}

func (mbdTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	// Force-Light: the variant arg is ignored. Every name resolves against
	// our own palette; anything unhandled falls through to the Fyne default
	// theme but always with VariantLight so we never accidentally inherit a
	// dark color when the host OS is in dark mode.
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
		return cOrange
	case theme.ColorNameForeground:
		return cFg
	case theme.ColorNameForegroundOnError:
		return cOnPrimary
	case theme.ColorNameForegroundOnPrimary:
		return cOnPrimary
	case theme.ColorNameForegroundOnSuccess:
		return cOnPrimary
	case theme.ColorNameForegroundOnWarning:
		return cFg
	case theme.ColorNameHeaderBackground:
		return cBackground
	case theme.ColorNameHover:
		return cHover
	case theme.ColorNameHyperlink:
		return cTeal
	case theme.ColorNameInputBackground:
		return cInputBg
	case theme.ColorNameInputBorder:
		return cInputBorder
	case theme.ColorNameMenuBackground:
		return cBackground
	case theme.ColorNameOverlayBackground:
		return cBackground
	case theme.ColorNamePlaceHolder:
		return cPlaceholder
	case theme.ColorNamePressed:
		return cPressed
	case theme.ColorNamePrimary:
		return cOrange
	case theme.ColorNameScrollBar:
		return cButton
	case theme.ColorNameScrollBarBackground:
		return color.NRGBA{0, 0, 0, 0}
	case theme.ColorNameSelection:
		return cOrangeSoft
	case theme.ColorNameSeparator:
		return cBorder
	case theme.ColorNameShadow:
		return cShadow
	case theme.ColorNameSuccess:
		return cTeal
	case theme.ColorNameWarning:
		return cOrange
	}
	return theme.DefaultTheme().Color(name, theme.VariantLight)
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
