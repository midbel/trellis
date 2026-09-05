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
