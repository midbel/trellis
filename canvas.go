package trellis

import (
	"fmt"
	"slices"
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

func (c Connector) Move(x, y int) Connector {
	cp := NewConnector(slices.Clone(c.Paths))
	for i := range cp.Paths {
		c.Paths[i].Start.X += x
		c.Paths[i].Start.Y += y
		c.Paths[i].End.X += x
		c.Paths[i].End.Y += y
	}
	return cp
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
	origin   Point
	dim      Dimension
	opts     *Options
	cells    []Placement
	children []*Canvas
}

func NewCanvas(opts *Options) (*Canvas, error) {
	canvas := &Canvas{
		dim: Dimension{
			Width:  opts.Width,
			Height: opts.Height,
		},
		opts: opts,
	}
	if err := canvas.dim.Validate(); err != nil {
		return nil, err
	}
	return canvas, nil
}

func (c *Canvas) SetOrigin(pt Point) {
	c.origin = pt
}

func (c *Canvas) Move(x, y int) {
	c.origin.X += x
	c.origin.Y += y
}

func (c *Canvas) Append(other *Canvas) {
	c.children = append(c.children, other)
}

func (c *Canvas) Screen() (View, error) {
	clone := c.cloneOptions()
	view, err := NewScreen(clone)
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
}

func (c *Canvas) Xml() (View, error) {
	clone := c.cloneOptions()
	view, err := NewXml(clone)
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
	// if n := len(c.children); n > 1 {
	// 	for _, cv := range c.children {
	// 		if err := view.put(cv.origin.X, cv.origin.Y, cv); err != nil {
	// 			return nil, err
	// 		}
	// 	}
	// 	return view, nil
	// } else {
	// 	if n == 0 {
	// 		return view, nil
	// 	}
	// 	return view, c.children[0].fillView(view)
	// }
}

func (c *Canvas) Svg() (View, error) {
	clone := c.cloneOptions()
	view, err := NewSvg(clone)
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
}

func (c *Canvas) Json() (View, error) {
	clone := c.cloneOptions()
	view, err := NewJson(clone)
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
}

func (c *Canvas) Table() (View, error) {
	return NewTable()
}

func (c *Canvas) cloneOptions() *Options {
	clone := c.opts.Clone()
	clone.Width = c.dim.Width
	clone.Height = c.dim.Height

	if len(c.children) > 0 {
		clone.ResetSize()
		if clone.Orient == HorizontalLayout {
			clone.Width = c.opts.Width
		} else {
			clone.Height = c.opts.Height
		}
		for _, x := range c.children {
			if clone.Orient == HorizontalLayout {
				clone.Height += x.dim.Height
				clone.Width = max(clone.Width, x.dim.Width)
			} else if clone.Orient == VerticalLayout {
				clone.Width += x.dim.Width
				clone.Height = max(clone.Height, x.dim.Height)
			}
		}
	}
	return clone
}

func (c *Canvas) fillView(view View) error {
	for _, p := range c.cells {
		p.X += c.origin.X
		p.Y += c.origin.Y
		if conn, ok := p.Cell.(Connector); ok {
			p.Cell = conn.Move(c.origin.X, c.origin.Y)
		}
		if err := view.put(p.X, p.Y, p.Cell); err != nil {
			return err
		}
	}
	for _, x := range c.children {
		if err := x.fillView(view); err != nil {
			return err
		}
	}
	return nil
}

func (c *Canvas) Resize(width, height int) error {
	c.dim.Width = width
	c.dim.Height = height
	return c.dim.Validate()
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
		return fmt.Errorf("invalid coordinate (%d, %d)")
	}
	p := Placement{
		Point: NewPoint(x, y),
		Cell:  cell,
	}
	c.cells = append(c.cells, p)
	return nil
}
