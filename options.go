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

func ParseFormat(str string) (Format, error) {
	return Regular, nil
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
	return AlignCenter, nil
}

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

func ParseConnector(str string) (ConnectorStyle, error) {
	return ConnectorAscii, nil
}

const (
	ConnectorAscii ConnectorStyle = iota
	ConnectorUnicode
)

type LayoutOptions struct {
	Width         int
	Height        int
	MinDepth      int
	MaxDepth      int
	HorizontalGap int
	VerticalGap   int
	Position      ParentPosition
	Reverse       bool
	Render        func(*Node, *Options) Content
}

type RenderOptions struct {
	ShowCoordinates bool
	Border          bool
}

type StyleOptions struct {
	Align       Alignment
	Padding     int
	PaddingChar string
	Style       ConnectorStyle
}

type Options struct {
	LayoutOptions
	RenderOptions
	StyleOptions
}

func prepareOptions(options *Options) *Options {
	opts := options
	if opts == nil {
		opts = defaultOptions.Clone()
	} else {
		opts = opts.Clone()
	}
	if opts.HorizontalGap == 0 {
		opts.HorizontalGap++
	}
	if opts.VerticalGap == 0 {
		opts.VerticalGap++
	}
	if opts.Render == nil {
		opts.Render = defaultRenderContent
	}
	return opts
}

func defaultRenderContent(node *Node, opts *Options) Content {
	value := []byte(node.Value)
	if opts.Padding > 0 {
		var (
			pad = make([]byte, opts.Padding)
			tmp = make([]byte, 0, len(value))
		)
		for i := range pad {
			pad[i] = ' '
		}
		tmp = append(tmp, pad...)
		tmp = append(tmp, value...)
		tmp = append(tmp, pad...)

		value = tmp
	}
	return Content{
		Value: value,
	}
}

var defaultOptions = &Options{
	LayoutOptions: LayoutOptions{
		VerticalGap:   DefaultVerticalGapSize,
		HorizontalGap: DefaultHorizontalGapSize,
		Position:      ParentAlignCenter,
	},
	StyleOptions: StyleOptions{
		Style:   ConnectorAscii,
		Padding: 1,
		Align:   AlignCenter,
	},
	RenderOptions: RenderOptions{
		Border: true,
	},
}

func (t *Options) Clone() *Options {
	x := *t
	return &x
}

func (t *Options) borderWidth() int {
	if t.Border {
		return 1
	}
	return 0
}

func (t *Options) paddedValue(str string) []byte {
	if t.Padding <= 0 {
		return []byte(str)
	}
	var (
		char  = byte(' ')
		value = []byte(str)
		pad   = make([]byte, t.Padding)
		tmp   = make([]byte, 0, t.Padding+len(str))
	)
	if t.PaddingChar != "" && len(t.PaddingChar) == 1 {
		char = t.PaddingChar[0]
	}
	for i := range pad {
		pad[i] = char
	}
	tmp = append(tmp, pad...)
	tmp = append(tmp, value...)
	tmp = append(tmp, pad...)
	return tmp
}

func (t *Options) hGaps() int {
	return t.HorizontalGap + t.HorizontalGap
}

func (t *Options) vGaps() int {
	return t.VerticalGap + t.VerticalGap
}
