package trellis

import (
	"unicode"

	"github.com/midbel/angle/svg"
)

type Content struct {
	Value []rune
	Style
}

func (c Content) String() string {
	return string(c.Value)
}

type StyleFlag uint8

const (
	StyleBold StyleFlag = 1 << iota
	StyleItalic
	StyleUnderline
)

type Style struct {
	Flags  StyleFlag
	Size   int
	Family string
	Color  string
}

func (s Style) Zero() bool {
	return s.Flags == 0 && s.Size == 0 && s.Family == "" && s.Color == ""
}

func (s Style) IsBold() bool {
	return s.Flags&StyleBold != 0
}

func (s Style) IsItalic() bool {
	return s.Flags&StyleItalic != 0
}

func (s Style) IsUnderline() bool {
	return s.Flags&StyleUnderline != 0
}

type Metrics interface {
	Width(str []rune) int
	Height() int
	Margin() int
}

type svgMetric struct {
	font   string
	size   int
	margin int
}

func (m svgMetric) Width(str []rune) int {
	z := svg.EstimateTextWidth(string(str), float64(m.size))
	return int(z)
}

func (m svgMetric) Height() int {
	return m.size
}

func (m svgMetric) Margin() int {
	return m.margin
}

type defaultMetric struct {
	margin int
}

func (defaultMetric) Width(str []rune) int {
	return DisplayWidth(str)
}

func (defaultMetric) Height() int {
	return 1
}

func (m defaultMetric) Margin() int {
	return m.margin
}

func DisplayWidth(value []rune) int {
	var width int
	for _, r := range value {
		width += RuneWidth(r)
	}
	return width
}

func RuneWidth(r rune) int {
	switch {
	case unicode.Is(unicode.Mn, r):
		return 0
	case unicode.Is(unicode.Me, r):
		return 0
	case unicode.Is(unicode.Cf, r):
		return 0
	default:
		return 1
	}
}
