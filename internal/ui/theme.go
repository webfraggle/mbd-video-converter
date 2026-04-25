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

// mbdTheme — "Departure Board (light)".
//
// A warm light variant of the industrial theme. Off-white paper background,
// a single deep-amber accent (#C77A1F — close to the orange of physical
// railway signals and old enamel station signs), type set on the Bigelow &
// Holmes Go fonts. The accent is desaturated against the light surface so
// it reads as authoritative rather than playful.
type mbdTheme struct{}

func newTheme() fyne.Theme { return mbdTheme{} }

// Palette
var (
	cBackground  = color.NRGBA{0xF7, 0xF4, 0xED, 0xFF} // warm off-white
	cInputBg     = color.NRGBA{0xFF, 0xFF, 0xFF, 0xFF}
	cButton      = color.NRGBA{0xEC, 0xE7, 0xDD, 0xFF}
	cHover       = color.NRGBA{0xE0, 0xD9, 0xCB, 0xFF}
	cPressed     = color.NRGBA{0xD4, 0xCC, 0xBC, 0xFF}
	cDisabledBtn = color.NRGBA{0xF0, 0xEB, 0xDF, 0xFF}
	cBorder      = color.NRGBA{0xDD, 0xD5, 0xC5, 0xFF}
	cFg          = color.NRGBA{0x1F, 0x1B, 0x16, 0xFF} // warm near-black
	cFgDisabled  = color.NRGBA{0xB5, 0xAF, 0xA3, 0xFF}
	cPlaceholder = color.NRGBA{0x95, 0x89, 0x7A, 0xFF}
	cAmber       = color.NRGBA{0xC7, 0x7A, 0x1F, 0xFF} // deep signal amber
	cAmberSoft   = color.NRGBA{0xC7, 0x7A, 0x1F, 0x26}
	cDanger      = color.NRGBA{0xB8, 0x3A, 0x3A, 0xFF}
	cShadow      = color.NRGBA{0x00, 0x00, 0x00, 0x20}
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

// White is used for text rendered on top of the amber primary fill (e.g. the
// HighImportance Convert and Save buttons). Plain off-white #FFFFFF.
var cOnAmber = color.NRGBA{0xFF, 0xFF, 0xFF, 0xFF}

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
		return cAmber
	case theme.ColorNameForeground:
		return cFg
	case theme.ColorNameForegroundOnError:
		return cOnAmber
	case theme.ColorNameForegroundOnPrimary:
		return cOnAmber
	case theme.ColorNameForegroundOnSuccess:
		return cOnAmber
	case theme.ColorNameForegroundOnWarning:
		return cFg
	case theme.ColorNameHeaderBackground:
		return cBackground
	case theme.ColorNameHover:
		return cHover
	case theme.ColorNameHyperlink:
		return cAmber
	case theme.ColorNameInputBackground:
		return cInputBg
	case theme.ColorNameInputBorder:
		return cBorder
	case theme.ColorNameMenuBackground:
		return cInputBg
	case theme.ColorNameOverlayBackground:
		return cBackground
	case theme.ColorNamePlaceHolder:
		return cPlaceholder
	case theme.ColorNamePressed:
		return cPressed
	case theme.ColorNamePrimary:
		return cAmber
	case theme.ColorNameScrollBar:
		return cButton
	case theme.ColorNameScrollBarBackground:
		return color.NRGBA{0, 0, 0, 0}
	case theme.ColorNameSelection:
		return cAmberSoft
	case theme.ColorNameSeparator:
		return cBorder
	case theme.ColorNameShadow:
		return cShadow
	case theme.ColorNameSuccess:
		return color.NRGBA{0x4E, 0x8C, 0x4E, 0xFF}
	case theme.ColorNameWarning:
		return color.NRGBA{0xC7, 0x7A, 0x1F, 0xFF}
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
