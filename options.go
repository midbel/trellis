package trellis

const (
	DefaultVerticalGapSize   = 2
	DefaultHorizontalGapSize = 5
)

const (
	connectBarAscii    = '+'
	verticalBarAscii   = '|'
	horizontalBarAscii = '-'
)

type Format uint8

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

const (
	AlignCenter Alignment = iota
	AlignLeft
	AlignRight
)

type ParentPosition uint8

const (
	ParentAlignCenter ParentPosition = iota
	ParentAlignFirst
	ParentAlignLast
)

type ConnectorStyle uint8

const (
	ConnectorAscii ConnectorStyle = iota
	ConnectorUnicode
)

type TreeRenderOptions struct {
	Width         int
	Height        int
	HorizontalGap int
	VerticalGap   int
	Padding       int
	Connector     ConnectorStyle
	Align         Alignment
	Position      ParentPosition
	Border        bool
}

var defaultTreeRenderOptions = &TreeRenderOptions{
	VerticalGap:   DefaultVerticalGapSize,
	HorizontalGap: DefaultHorizontalGapSize,
	Padding:       1,
	Connector:     ConnectorAscii,
	Align:         AlignCenter,
	Position:      ParentAlignCenter,
	Border:        true,
}

func (t *TreeRenderOptions) clone() *TreeRenderOptions {
	x := *t
	return &x
}
