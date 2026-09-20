package trellis

import (
	"unicode"
)

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
