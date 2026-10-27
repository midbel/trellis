package trellis

import (
	"fmt"
	"io"
)

func Render(w io.Writer, root *Node, rdr *RenderOptions, options *OutputOptions) error {
	if options == nil {
		return fmt.Errorf("options should be provided")
	}

	fn, ok := renderers[rdr.Orient]
	if !ok {
		return fmt.Errorf("unsupported layout given")
	}
	items, err := prepareRender(root, rdr, options)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	canvas, err := fn(items, rdr)
	if err != nil {
		return err
	}
	return canvas.Render(w, *options)
}

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

	opts RenderOptions
}

func NewCanvas(size Dimension, opts *RenderOptions) (*Canvas, error) {
	canvas := &Canvas{
		dim:  size,
		opts: *opts,
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

func (c *Canvas) Render(w io.Writer, opts OutputOptions) error {
	var view View
	switch opts.Mode {
	case OutputSvg:
		opts, err := opts.toSvg()
		if err != nil {
			return err
		}
		opts.orient = c.opts.Orient
		view, err = NewSvg(c.adjustSize(), opts)
		if err != nil {
			return err
		}
	case OutputScreen:
		opts, err := opts.toScreen()
		if err != nil {
			return err
		}
		opts.orient = c.opts.Orient
		view, err = NewScreen(c.adjustSize(), opts)
		if err != nil {
			return err
		}
	case OutputXml:
		opts, err := opts.toXml()
		if err != nil {
			return err
		}
		opts.orient = c.opts.Orient
		view, err = NewXml(c.adjustSize(), opts)
		if err != nil {
			return err
		}
	case OutputJson:
		opts, err := opts.toJson()
		if err != nil {
			return err
		}
		opts.orient = c.opts.Orient
		view, err = NewJson(c.adjustSize(), opts)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported view type")
	}
	if err := c.fillView(view); err != nil {
		return err
	}
	return view.Render(w)
}

func (c *Canvas) Table(opts TableOptions) (View, error) {
	view, err := NewTable()
	if err != nil {
		return nil, err
	}
	return view, c.fillView(view)
}

func (c *Canvas) adjustSize() Dimension {
	z := c.dim
	if len(c.children) == 0 {
		return z
	}
	z.Reset()

	if c.opts.Orient == HorizontalLayout || c.opts.Orient == CompactLayout {
		z.Width = c.dim.Width
	} else {
		z.Height = c.dim.Height
	}
	for _, x := range c.children {
		if c.opts.Orient == HorizontalLayout || c.opts.Orient == CompactLayout {
			z.Height += x.dim.Height
			z.Width = max(z.Width, x.dim.Width)
		} else if c.opts.Orient == VerticalLayout {
			z.Width += x.dim.Width
			z.Height = max(z.Height, x.dim.Height)
		}
	}
	return z
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

var renderers = map[Orientation]func([]*Item, *RenderOptions) (*Canvas, error){
	HorizontalLayout: buildHorizontal,
	VerticalLayout:   buildVertical,
	CompactLayout:    buildCompact,
}

func buildHorizontal(items []*Item, opts *RenderOptions) (*Canvas, error) {
	sizes, err := allocateFromSize(opts.Size, items, opts.AllocateMode, HorizontalLayout)
	if err != nil {
		return nil, err
	}
	if opts.Size, err = adjustSizes(sizes, HorizontalLayout); err != nil {
		return nil, err
	}
	master, err := NewCanvas(opts.Size, opts)
	if err != nil {
		return nil, err
	}
	var offset int
	for i := range items {
		set, err := stdHorizontalLayout(items[i], cloneAndResizeOptions(*opts, sizes[i]))
		if err != nil {
			return nil, err
		}
		canvas, err := NewCanvas(set.Dimension(), opts)
		if err != nil {
			return nil, err
		}
		for _, i := range set.Items {
			if err := canvas.Put(i.X(), i.Y(), i.Content); err != nil {
				return nil, err
			}
			for _, x := range i.Children {
				path := horizontalPath(i, x)
				if err := canvas.Put(path.X(), path.Y(), path); err != nil {
					return nil, err
				}
			}
		}
		canvas.Move(0, offset)
		master.Append(canvas)
		offset += set.Height
	}
	return master, nil
}

func buildVertical(items []*Item, opts *RenderOptions) (*Canvas, error) {
	sizes, err := allocateFromSize(opts.Size, items, opts.AllocateMode, VerticalLayout)
	if err != nil {
		return nil, err
	}
	if opts.Size, err = adjustSizes(sizes, VerticalLayout); err != nil {
		return nil, err
	}
	master, err := NewCanvas(opts.Size, opts)
	if err != nil {
		return nil, err
	}
	var offset int
	for i := range items {
		set, err := stdVerticalLayout(items[i], cloneAndResizeOptions(*opts, sizes[i]))
		if err != nil {
			return nil, err
		}

		canvas, err := NewCanvas(set.Dimension(), opts)
		if err != nil {
			return nil, err
		}
		for _, i := range set.Items {
			if err := canvas.Put(i.X(), i.Y(), i.Content); err != nil {
				return nil, err
			}
			for _, x := range i.Children {
				path := verticalPath(i, x)
				if err := canvas.Put(path.X(), path.Y(), path); err != nil {
					return nil, err
				}
			}
		}
		canvas.Move(offset, 0)
		master.Append(canvas)
		offset += set.Width
	}
	return master, nil
}

func buildCompact(items []*Item, opts *RenderOptions) (*Canvas, error) {
	opts.Placement = AlignStart
	master, err := NewCanvas(opts.Size, opts)
	if err != nil {
		return nil, err
	}
	var (
		offset  int
		recurse func(*Canvas, *Item, int) error
	)

	recurse = func(canvas *Canvas, i *Item, parent int) error {
		if err := canvas.Put(i.Position.X, i.Position.Y, i.Content); err != nil {
			return err
		}
		h1 := NewPoint(parent, i.Position.Y)
		h2 := NewPoint(i.Position.X-1, i.Position.Y)
		h := NewPath(h1, h2)
		if h.DistanceX() > 0 {
			if err := canvas.Put(h.X(), h.Y(), h); err != nil {
				return err
			}
		}
		if i.Leaf() {
			return nil
		}
		v1 := NewPoint(parent, i.Position.Y+1)
		v2 := NewPoint(parent, i.Position.Y+i.Weight())
		v := NewPath(v1, v2)
		if v.DistanceY() > 0 {
			if err := canvas.Put(v.X(), v.Y(), v); err != nil {
				return err
			}
		}

		for _, c := range i.Children {
			if err := recurse(canvas, c, i.Position.X); err != nil {
				return err
			}
		}
		return nil
	}
	for i := range items {
		clone := opts.Clone()

		set, err := compactLayout(items[i], clone)
		if err != nil {
			return nil, err
		}
		canvas, err := NewCanvas(set.Dimension(), opts)
		if err != nil {
			return nil, err
		}
		root, err := set.Root()
		if err != nil {
			return nil, err
		}
		if err := recurse(canvas, root, 0); err != nil {
			return nil, err
		}
		canvas.Move(0, offset)
		master.Append(canvas)
		offset += set.Height
	}
	return master, nil
}

func sunburst(root *Node, options Options) error {
	return fmt.Errorf("not yet implemented")
}

func radial(w io.Writer, root *Node, options Options) error {
	return fmt.Errorf("not yet implemented")
}

func treeMap(w io.Writer, root *Node, options Options) error {
	return fmt.Errorf("not yet implemented")
}

func cloneAndResizeOptions(opts RenderOptions, size Dimension) RenderOptions {
	clone := opts.Clone()
	clone.Size = size
	return clone
}

func prepareRender(node *Node, opts *RenderOptions, options *OutputOptions) ([]*Item, error) {
	if node == nil {
		return nil, fmt.Errorf("root node should be given")
	}
	if opts.Orient == HorizontalLayout {
		opts.estimateMinSize = estimateMinSize(node, opts.Size.Height)
	} else {
		opts.estimateMinSize = estimateMinSize(node, opts.Size.Width)
	}
	return traverseNodes(traverse(node, opts.MinDepth), opts, options)
}

func traverseNodes(nodes []*Node, opts *RenderOptions, options *OutputOptions) ([]*Item, error) {
	list := make([]*Item, 0, len(nodes))
	for _, n := range nodes {
		i, err := transformNode(n, opts, options)
		if err != nil {
			return nil, err
		}
		i.root = true
		list = append(list, i)
	}
	return list, nil
}

func transformNode(n *Node, opts *RenderOptions, options *OutputOptions) (*Item, error) {
	it := &Item{
		Content: n.Content(),
		Metric:  options.Metrics(n),
	}
	for _, n := range n.Nodes {
		sub, err := transformNode(n, opts, options)
		if err != nil {
			return nil, err
		}
		it.Children = append(it.Children, sub)
	}
	return it, nil
}