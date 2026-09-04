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

type Style struct {
	Bold      bool
	Italic    bool
	Underline bool
}

type Content struct {
	Value     []byte
	Style

	kind ContentKind
}

func (c Content) String() string {
	return string(c.Value)
}

type Cell interface {
	Byte() byte
}

type Canvas struct {
	dim   Dimension
	cells []Cell
}

func NewCanvas(width, height int) *Canvas {
	return &Canvas{
		dim: Dimension{
			Width:  width,
			Height: height,
		},
		cells: make([]Cell, width*height),
	}
}

func (c *Canvas) Put(x, y int, content Content) {
	for i, b := range content.Value {
		c.put(x+i, y, b)
	}
}

func (c *Canvas) Connect(seg Segment) {
	if seg.Horizontal() {
		c.horizontalConnector(seg)
	} else {
		c.verticalConnector(seg)
	}
}

func (c *Canvas) Render(sc *Screen) error {
	for i, ct := range c.cells {
		y := i / c.dim.Width
		x := i % c.dim.Width

		var ch byte
		if ct != nil {
			ch = ct.Byte()
		}
		sc.Put(x, y, ch)
	}
	return nil
}

func (c *Canvas) put(x, y int, ch byte) {
	if !c.dim.Valid(x, y) {
		return
	}
	c.cells[y*c.dim.Width+x] = newChar(ch)
}

func (c *Canvas) putConnector(x, y int, ch byte) {
	if !c.dim.Valid(x, y) {
		return
	}
	cell := c.cells[y*c.dim.Width+x]
	if cell != nil {
		b := cell.Byte()
		if b == connectBarAscii && ch == connectBarAscii {
			return
		}
		if b == verticalBarAscii && ch == horizontalBarAscii {
			ch = connectBarAscii
		} else if b == horizontalBarAscii && ch == verticalBarAscii {
			ch = connectBarAscii
		}
	}
	c.put(x, y, ch)
}

func (c *Canvas) verticalConnector(seg Segment) {
	var (
		start = seg.Start
		end   = seg.End
	)
	if start.BeforeY(end) {
		start, end = end, start
	}
	for y := start.Y; y >= end.Y; y-- {
		ch := verticalBarAscii
		if y == start.Y || y == end.Y {
			ch = connectBarAscii
		}
		c.putConnector(start.X, y, ch)
	}
}

func (c *Canvas) horizontalConnector(seg Segment) {
	var (
		start = seg.Start
		end   = seg.End
	)
	if end.BeforeX(start) {
		start, end = end, start
	}
	for x := start.X; x <= end.X; x++ {
		ch := horizontalBarAscii
		if x == start.X || x == end.X {
			ch = connectBarAscii
		}
		c.putConnector(x, start.Y, ch)
	}
}

type char struct {
	value byte
}

func newChar(b byte) Cell {
	return char{
		value: b,
	}
}

func (c char) Byte() byte {
	return c.value
}
