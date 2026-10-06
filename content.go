package trellis

import (
	"unicode"
)

type Content struct {
	Value []rune
	Style
}

func (c Content) String() string {
	return string(c.Value)
}

func (c Content) DisplayWidth() int {
	return DisplayWidth(c.Value)
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
}

type defaultMetric struct{}

func (defaultMetric) Width(str []rune) int {
	return DisplayWidth(str)
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