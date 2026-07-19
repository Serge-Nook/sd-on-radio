package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"github.com/Serge-Nook/sd-on-radio/assets"
)

// deckTheme is a dark, touch-friendly theme with Cyrillic-capable fonts and an
// orange accent, tuned for the Steam Deck screen.
type deckTheme struct {
	regular fyne.Resource
	bold    fyne.Resource
}

func newDeckTheme() *deckTheme {
	return &deckTheme{
		regular: fyne.NewStaticResource("DejaVuSans.ttf", assets.FontRegular),
		bold:    fyne.NewStaticResource("DejaVuSans-Bold.ttf", assets.FontBold),
	}
}

func (t *deckTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Bold {
		return t.bold
	}
	return t.regular
}

func (t *deckTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0xff, G: 0x69, B: 0x42, A: 0xff}
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x14, G: 0x19, B: 0x23, A: 0xff}
	case theme.ColorNameButton, theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x22, G: 0x29, B: 0x36, A: 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xf2, G: 0xf4, B: 0xf8, A: 0xff}
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (t *deckTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *deckTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 16
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 10
	}
	return theme.DefaultTheme().Size(name)
}
