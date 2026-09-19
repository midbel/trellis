package trellis

import (
	"fmt"
	"strings"
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

type Dimension struct {
	Width  int
	Height int
}

func (d Dimension) Validate() error {
	if d.Width <= 0 {
		return fmt.Errorf("width can not be equal to 0 or negative")
	}
	if d.Height <= 0 {
		return fmt.Errorf("height can not be equal to 0 or negative")
	}
	return nil
}

func (d Dimension) Valid(x, y int) bool {
	return x >= 0 && x < d.Width && y >= 0 && y < d.Height
}

type Style struct {
	Bold      bool
	Italic    bool
	Underline bool
}

type Cell interface{}

type Connector struct {
	Paths []Segment
}

func NewConnector(paths []Segment) Connector {
	return Connector{
		Paths: paths,
	}
}

func (c Connector) X() int {
	return c.Paths[0].Start.X
}

func (c Connector) Y() int {
	return c.Paths[0].Start.Y
}

func (c Connector) String() string {
	var str strings.Builder
	str.WriteString("connector(")
	for i, p := range c.Paths {
		if i > 0 {
			str.WriteRune(',')
			str.WriteRune(' ')
		}
		str.WriteString(p.String())
	}
	str.WriteString(")")
	return str.String()
}

type Content struct {
	Value []rune
	Style
}

func (c Content) String() string {
	return string(c.Value)
}

func (c Content) DisplayWidth() int {
	var width int
	for _, r := range c.Value {
		width += RuneWidth(r)
	}
	return width
}

type Canvas struct {
	dim   Dimension
	cells []Cell
}

func NewCanvas(width, height int) (*Canvas, error) {
	canvas := &Canvas{
		dim: Dimension{
			Width:  width,
			Height: height,
		},
		cells: make([]Cell, width*height),
	}
	if err := canvas.dim.Validate(); err != nil {
		return nil, err
	}
	return canvas, nil
}

func (c *Canvas) Put(x, y int, cell Cell) error {
	if conn, ok := cell.(Connector); ok {
		return c.PutConnector(conn)
	}
	return c.put(x, y, cell)
}

func (c *Canvas) PutConnector(conn Connector) error {
	for _, s := range conn.Paths {
		err := c.put(s.Start.X, s.Start.Y, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Canvas) VerticalBar(x, y, size int, halfOpen bool) error {
	var (
		beg = NewPoint(x, y)
		end = NewPoint(x, y+size)
		seg = NewSegment(beg, end)
	)
	return c.put(x, y, NewConnector([]Segment{seg}))
}

func (c *Canvas) HorizontalBar(x, y, size int, halfOpen bool) error {
	var (
		beg = NewPoint(x, y)
		end = NewPoint(x+size, y)
		seg = NewSegment(beg, end)
	)
	return c.put(x, y, NewConnector([]Segment{seg}))
}

func (c *Canvas) Render(sc *Screen) error {
	for i, cell := range c.cells {
		y := i / c.dim.Width
		x := i % c.dim.Width

		if err := sc.Put(x, y, cell); err != nil {
			return err
		}
	}
	return nil
}

func (c *Canvas) put(x, y int, cell Cell) error {
	if !c.dim.Valid(x, y) {
		return fmt.Errorf("invalid coordinate (%d, %d)", x, y)
	}
	c.cells[y*c.dim.Width+x] = cell
	return nil
}
