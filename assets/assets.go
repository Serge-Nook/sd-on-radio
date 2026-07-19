// Package assets embeds application resources (icon and fonts).
package assets

import _ "embed"

// IconSVG is the application icon in SVG form.
//
//go:embed icon.svg
var IconSVG []byte

// FontRegular is a Cyrillic-capable regular font (DejaVu Sans).
//
//go:embed DejaVuSans.ttf
var FontRegular []byte

// FontBold is a Cyrillic-capable bold font (DejaVu Sans Bold).
//
//go:embed DejaVuSans-Bold.ttf
var FontBold []byte
