package trellis

import (
	"fmt"
	"strings"
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

type Placement struct {
	Point
	Cell
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
	return DisplayWidth(c.Value)
}

type Canvas struct {
	dim   Dimension
	opts  *Options
	cells []Placement
}

func NewCanvas(opts *Options) (*Canvas, error) {
	canvas := &Canvas{
		dim: Dimension{
			Width:  opts.Width,
			Height: opts.Height,
		},
		opts:  opts,
		cells: make([]Placement, 0, opts.Width*opts.Height),
	}
	if err := canvas.dim.Validate(); err != nil {
		return nil, err
	}
	return canvas, nil
}

func (c *Canvas) Screen() (View, error) {
	view, err := NewScreen(c.opts)
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
}

func (c *Canvas) Xml() (View, error) {
	view, err := NewXml(c.opts)
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
}

func (c *Canvas) Svg() (View, error) {
	return nil, fmt.Errorf("svg view: not yet implemented")
}

func (c *Canvas) Json() (View, error) {
	return nil, fmt.Errorf("json view: not yet implemented")
}

func (c *Canvas) fillView(view View) error {
	for _, p := range c.cells {
		if err := view.put(p.X, p.Y, p.Cell); err != nil {
			return err
		}
	}
	return nil
}

func (c *Canvas) Put(x, y int, cell Cell) error {
	return c.put(x, y, cell)
}

func (c *Canvas) VerticalBar(x, y, size int) error {
	var (
		beg = NewPoint(x, y)
		end = NewPoint(x, y+size)
		seg = NewSegment(beg, end)
	)
	return c.put(x, y, NewConnector([]Segment{seg}))
}

func (c *Canvas) HorizontalBar(x, y, size int) error {
	var (
		beg = NewPoint(x, y)
		end = NewPoint(x+size, y)
		seg = NewSegment(beg, end)
	)
	return c.put(x, y, NewConnector([]Segment{seg}))
}

func (c *Canvas) put(x, y int, cell Cell) error {
	if !c.dim.Valid(x, y) {
		return fmt.Errorf("invalid coordinate (%d, %d)", x, y)
	}
	p := Placement{
		Point: NewPoint(x, y),
		Cell:  cell,
	}
	c.cells = append(c.cells, p)
	return nil
}
