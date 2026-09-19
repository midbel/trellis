package trellis

import "fmt"

type Orientation uint8

func ParseOrientation(orient string) (Orientation, error) {
	switch orient {
	case "", "h", "horizontal":
		return HorizontalLayout, nil
	case "v", "vertical":
		return VerticalLayout, nil
	case "c", "compact":
		return CompactLayout, nil
	default:
		return HorizontalLayout, fmt.Errorf("%s: unknown orientation", orient)
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

const (
	HorizontalLayout Orientation = iota
	VerticalLayout
	CompactLayout
)

type Format uint8

func ParseFormat(str string) (Format, error) {
	switch str {
	case "", "regular":
		return Regular, nil
	case "italic":
		return Italic, nil
	case "bold":
		return Bold, nil
	case "underline":
		return Underline, nil
	case "strike":
		return Strike, nil
	default:
		return 0, unknown("format", str)
	}
}

const (
	Regular Format = 1 << iota
	Italic
	Bold
	Underline
	Strike
)

func (f Format) Zero() bool {
	return f <= Regular
}

type Alignment uint8

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
	ConnectorAscii ConnectorStyle = iota
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
