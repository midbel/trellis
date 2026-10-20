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
	Flags   StyleFlag
	Size    int
	Family  string
	Color   string
	Border  bool
	Margin  Insets
	Padding Insets
	AlignX  Alignment
	AlignY  Alignment
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

func hasStyleForScreen(s Style) bool {
	return s.Flags != 0 || s.Color != ""
}

type Metrics interface {
	Width(str []rune) int
	Height() int
}

type Output uint8

const (
	OutputScreen = 1 << iota
	OutputSvg
	OutputXml
	OutputJson
	OutputTable
)

func ParseOutput(str string) (Output, error) {
	switch str {
	case "screen", "terminal":
		return OutputScreen, nil
	case "svg":
		return OutputSvg, nil
	case "xml":
		return OutputXml, nil
	case "json":
		return OutputJson, nil
	case "inspect", "table", "debug":
		return OutputTable, nil
	default:
		return 0, unknown("output", str)
	}
}

type Orientation uint8

const (
	HorizontalLayout Orientation = 1 << iota
	VerticalLayout
	CompactLayout
)

func ParseOrientation(str string) (Orientation, error) {
	switch str {
	case "", "h", "horizontal":
		return HorizontalLayout, nil
	case "v", "vertical":
		return VerticalLayout, nil
	case "c", "compact":
		return CompactLayout, nil
	default:
		return 0, unknown("orientation", str)
	}
}

func (o Orientation) String() string {
	switch o {
	case HorizontalLayout:
		return "horizontal"
	case VerticalLayout:
		return "vertical"
	case CompactLayout:
		return "compact"
	default:
		return ""
	}
}

type Alignment uint8

const (
	AlignCenter Alignment = iota
	AlignStart
	AlignEnd
)

const (
	AlignLeft   = AlignStart
	AlignTop    = AlignStart
	AlignRight  = AlignEnd
	AlignBottom = AlignEnd
)

func ParseAlignment(str string) (Alignment, error) {
	switch str {
	case "", "center":
		return AlignCenter, nil
	case "left", "start", "top":
		return AlignStart, nil
	case "right", "end", "bottom":
		return AlignEnd, nil
	default:
		return 0, unknown("alignment", str)
	}
}

const (
	connectBarAscii    = '+'
	verticalBarAscii   = '|'
	horizontalBarAscii = '-'
)

const (
	verticalBarUnicode    = '│'
	horizontalBarUnicode  = '─'
	crossingUnicode       = '┼'
	downRightUnicode      = '┌'
	downLeftUnicode       = '┐'
	upRightUnicode        = '└'
	upLeftUnicode         = '┘'
	horizontalDownUnicode = '┬'
	verticalRightUnicode  = '├'
	verticalLeftUnicode   = '┤'
	horizontalTopUnicode  = '┴'
)

func IsBar(r rune) bool {
	switch r {
	default:
		return false
	case connectBarAscii:
	case verticalBarAscii:
	case horizontalBarAscii:
	case verticalBarUnicode:
	case horizontalBarUnicode:
	case crossingUnicode:
	case downRightUnicode:
	case downLeftUnicode:
	case upRightUnicode:
	case upLeftUnicode:
	case horizontalDownUnicode:
	case verticalRightUnicode:
	case verticalLeftUnicode:
	case horizontalTopUnicode:
	}
	return true
}

type ConnectorStyle uint8

func ParseConnector(str string) (ConnectorStyle, error) {
	switch str {
	case "", "ascii", "classic":
		return ConnectorAscii, nil
	case "unicode":
		return ConnectorUnicode, nil
	default:
		return 0, unknown("connector", str)
	}
}

const (
	ConnectorAscii ConnectorStyle = 1 << iota
	ConnectorUnicode
)

func (c ConnectorStyle) CrossPath() rune {
	if c == ConnectorAscii {
		return connectBarAscii
	}
	return crossingUnicode
}

func (c ConnectorStyle) VerticalLeft() rune {
	if c == ConnectorAscii {
		return connectBarAscii
	}
	return verticalLeftUnicode
}

func (c ConnectorStyle) VerticalRight() rune {
	if c == ConnectorAscii {
		return connectBarAscii
	}
	return verticalRightUnicode
}

func (c ConnectorStyle) HorizontalUp() rune {
	if c == ConnectorAscii {
		return connectBarAscii
	}
	return horizontalTopUnicode
}

func (c ConnectorStyle) HorizontalDown() rune {
	if c == ConnectorAscii {
		return connectBarAscii
	}
	return horizontalDownUnicode
}

func (c ConnectorStyle) TopLeft() rune {
	if c == ConnectorAscii {
		return connectBarAscii
	}
	return downRightUnicode
}

func (c ConnectorStyle) TopRight() rune {
	if c == ConnectorAscii {
		return connectBarAscii
	}
	return downLeftUnicode
}

func (c ConnectorStyle) BottomLeft() rune {
	if c == ConnectorAscii {
		return connectBarAscii
	}
	return upRightUnicode
}

func (c ConnectorStyle) BottomRight() rune {
	if c == ConnectorAscii {
		return connectBarAscii
	}
	return upLeftUnicode
}

func (c ConnectorStyle) VerticalBar() rune {
	if c == ConnectorAscii {
		return verticalBarAscii
	}
	return verticalBarUnicode
}

func (c ConnectorStyle) HorizontalBar() rune {
	if c == ConnectorAscii {
		return horizontalBarAscii
	}
	return horizontalBarUnicode
}

type svgMetric struct {
	font string
	size int
}

func (m svgMetric) Width(str []rune) int {
	z := svg.EstimateTextWidthForFont(string(str), m.font, float64(m.size))
	return int(z)
}

func (m svgMetric) Height() int {
	return m.size
}

type defaultMetric struct{}

func (defaultMetric) Width(str []rune) int {
	return DisplayWidth(str)
}

func (defaultMetric) Height() int {
	return 1
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
