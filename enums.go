package trellis

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
