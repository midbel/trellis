package trellis

import (
	"fmt"
)

type Placement struct {
	Point
	Cell
}

type Cell interface{}

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

func (c *Canvas) Move(x, y int) {
	c.origin.X += x
	c.origin.Y += y
}

func (c *Canvas) Append(other *Canvas) {
	c.children = append(c.children, other)
}

func (c *Canvas) Put(x, y int, cell Cell) error {
	return c.put(x, y, cell)
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
		if path, ok := p.Cell.(Path); ok {
			p.Cell = path.Move(c.origin.X, c.origin.Y)
		}
		if err := d.draw(p.X, p.Y, p.Cell); err != nil {
			return err
		}
	}
	return nil
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
