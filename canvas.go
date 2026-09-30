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

func NewDimension(w int, h int) Dimension {
	d := Dimension{
		Width:  w,
		Height: h,
	}
	return d
}

func (d *Dimension) Validate() error {
	if d.Width <= 0 {
		return fmt.Errorf("width can not be equal to 0 or negative")
	}
	if d.Height <= 0 {
		return fmt.Errorf("height can not be equal to 0 or negative")
	}
	return nil
}

func (d *Dimension) Resize(w, h int) error {
	d.Width = w
	d.Height = h
	return d.Validate()
}

func (d *Dimension) Valid(x, y int) bool {
	return x >= 0 && x < d.Width && y >= 0 && y < d.Height
}

func adjustSizes(sizes []Dimension, orient Orientation) (Dimension, error) {
	var dim Dimension
	switch orient {
	case HorizontalLayout:
		for i := range sizes {
			dim.Width = max(sizes[i].Width, dim.Width)
			dim.Height += sizes[i].Height
		}
	case VerticalLayout:
		for i := range sizes {
			dim.Height = max(sizes[i].Height, dim.Height)
			dim.Width += sizes[i].Width
		}
	default:
		return dim, fmt.Errorf("unsupported orientation")
	}
	return dim, nil
}

func estimateMinSize(node *Node, size int) int {
	count := node.Count()
	if count == 0 {
		return size
	}
	tmp := size / count
	if mod := size % count; mod != 0 {
		tmp += mod
	}
	return tmp
}

func allocateFromSize(size Dimension, nodes []*Node, mode Allocate, orient Orientation) ([]Dimension, error) {
	if len(nodes) == 1 {
		return []Dimension{size}, nil
	}
	list := make([]Dimension, 0, len(nodes))
	switch mode {
	case AllocateEqual:
		if orient == HorizontalLayout {
			height := size.Height / len(nodes)
			for range nodes {
				x := NewDimension(size.Width, height)
				list = append(list, x)
			}
		} else if orient == VerticalLayout {
			width := size.Width / len(nodes)
			for range nodes {
				x := NewDimension(width, size.Height)
				list = append(list, x)
			}
		} else {
			return nil, fmt.Errorf("unsupported orientation")
		}
	case AllocateProportional:
		var total int
		for i := range nodes {
			total += nodes[i].Weight()
		}
		if orient == HorizontalLayout {
			for i := range nodes {
				h := size.Height * nodes[i].Weight() / total
				x := NewDimension(size.Width, h)
				list = append(list, x)
			}
		} else if orient == VerticalLayout {
			for i := range nodes {
				w := size.Width * nodes[i].Weight() / total
				x := NewDimension(w, size.Height)
				list = append(list, x)
			}
		} else {
			return nil, fmt.Errorf("unsupported orientation")
		}
	default:
		return nil, fmt.Errorf("unsupported allocation mode")
	}
	return list, nil
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
		cp.Paths[i].Start.X += x
		cp.Paths[i].Start.Y += y
		cp.Paths[i].End.X += x
		cp.Paths[i].End.Y += y
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
	cells    []Placement
	children []*Canvas
}

func NewCanvas(size Dimension) (*Canvas, error) {
	canvas := &Canvas{
		dim: size,
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

func (c *Canvas) Screen(opts ScreenOptions) (View, error) {
	opts.RenderOptions = c.adjustSize(opts.RenderOptions)
	view, err := NewScreen(opts)
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
}

func (c *Canvas) Xml(opts XmlOptions) (View, error) {
	opts.RenderOptions = c.adjustSize(opts.RenderOptions)
	view, err := NewXml(opts)
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
}

func (c *Canvas) Svg(opts SvgOptions) (View, error) {
	opts.RenderOptions = c.adjustSize(opts.RenderOptions)
	view, err := NewSvg(opts)
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
}

func (c *Canvas) Json(opts JsonOptions) (View, error) {
	opts.RenderOptions = c.adjustSize(opts.RenderOptions)
	view, err := NewJson(opts)
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
}

func (c *Canvas) Table(opts TableOptions) (View, error) {
	view, err := NewTable()
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
}

func (c *Canvas) adjustSize(opts RenderOptions) RenderOptions {
	if len(c.children) > 0 {
		opts.ResetSize()
		if opts.Orient == HorizontalLayout {
			opts.Size.Width = c.dim.Width
		} else {
			opts.Size.Height = c.dim.Height
		}
		for _, x := range c.children {
			if opts.Orient == HorizontalLayout {
				opts.Size.Height += x.dim.Height
				opts.Size.Width = max(opts.Size.Width, x.dim.Width)
			} else if opts.Orient == VerticalLayout {
				opts.Size.Width += x.dim.Width
				opts.Size.Height = max(opts.Size.Height, x.dim.Height)
			}
		}
	}
	return opts
}

func (c *Canvas) fillView(view View) error {
	if err := c.drawCells(view); err != nil {
		return err
	}
	for _, x := range c.children {
		err := view.fill(x.origin.X, x.origin.Y, func(draw drawer) error {
			return x.drawCells(draw)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Canvas) drawCells(d drawer) error {
	for _, p := range c.cells {
		p.X += c.origin.X
		p.Y += c.origin.Y
		if conn, ok := p.Cell.(Connector); ok {
			p.Cell = conn.Move(c.origin.X, c.origin.Y)
		}
		if err := d.draw(p.X, p.Y, p.Cell); err != nil {
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
		return fmt.Errorf("invalid coordinate (%d, %d)", x, y)
	}
	p := Placement{
		Point: NewPoint(x, y),
		Cell:  cell,
	}
	c.cells = append(c.cells, p)
	return nil
}
