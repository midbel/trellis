package trellis

const (
	connectBarAscii    byte = '+'
	verticalBarAscii   byte = '|'
	horizontalBarAscii byte = '-'
)

type Dimension struct {
	Width  int
	Height int
}

func (d Dimension) Valid(x, y int) bool {
	return x >= 0 && x < d.Width && y >= 0 && y < d.Height
}

type ContentKind uint8

const (
	KindValue ContentKind = iota
	KindConnector
)

type Content struct {
	Value     []byte
	Bold      bool
	Italic    bool
	Underline bool

	kind ContentKind
}

func (c Content) String() string {
	return string(c.Value)
}

func Connector(segments []Segment) []*Item {
	var (
		list []*Item
		tmp  []*Item
	)
	for _, s := range segments {
		if s.Start.Y == s.End.Y {
			tmp = horizontalConnector(s)
		} else {
			tmp = verticalConnector(s)
		}
		list = append(list, tmp...)
	}
	return list
}

func verticalConnector(seg Segment) []*Item {
	var (
		list  []*Item
		start = seg.Start
		end   = seg.End
	)
	if start.BeforeY(end) {
		start, end = end, start
	}
	for y := start.Y; y >= end.Y; y-- {
		char := verticalBarAscii
		if y == start.Y || y == end.Y {
			char = connectBarAscii
		}
		content := Content{
			Value: []byte{char},
			kind:  KindConnector,
		}
		i := &Item{
			Content: content,
			Point: Point{
				X: start.X,
				Y: y,
			},
		}
		list = append(list, i)
	}
	return list
}

func horizontalConnector(seg Segment) []*Item {
	var (
		list  []*Item
		start = seg.Start
		end   = seg.End
	)
	if end.BeforeX(start) {
		start, end = end, start
	}
	for x := start.X; x <= end.X; x++ {
		char := horizontalBarAscii
		if x == start.X || x == end.X {
			char = connectBarAscii
		}
		content := Content{
			Value: []byte{char},
			kind:  KindConnector,
		}
		i := &Item{
			Content: content,
			Point: Point{
				X: x,
				Y: start.Y,
			},
		}
		list = append(list, i)
	}
	return list
}

type Canvas struct {
	dim   Dimension
	cells []Content
}

func NewCanvas(width, height int) *Canvas {
	return &Canvas{
		dim: Dimension{
			Width:  width,
			Height: height,
		},
		cells: make([]Content, width*height),
	}
}

func (c *Canvas) Put(x, y int, content Content) error {
	if !c.dim.Valid(x, y) {
		return nil
	}
	p := c.cells[y*c.dim.Width+x]
	if p.kind == KindConnector && content.kind == KindConnector {
		var (
			source = p.Value[0]
			target = content.Value[0]
		)
		if source == connectBarAscii {
			return nil
		}
		content.Value[0] = replaceConnector(source, target)
	}
	c.cells[y*c.dim.Width+x] = content
	return nil
}

func (c *Canvas) Render(sc *Screen) error {
	for i, ct := range c.cells {
		y := i / c.dim.Width
		x := i % c.dim.Width
		sc.Put(x, y, ct)
	}
	return nil
}

func replaceConnector(source, target byte) byte {
	switch {
	case source == connectBarAscii || target == connectBarAscii:
		return connectBarAscii
	case source == verticalBarAscii && target == horizontalBarAscii:
		return connectBarAscii
	case source == horizontalBarAscii && target == verticalBarAscii:
		return connectBarAscii
	default:
		return target
	}
}
