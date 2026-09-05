package trellis

import (
	"io"
)

type Renderer interface {
	Render(*Node, *Options) error
}

func HorizontalTree(w io.Writer, root *Node, opts *Options) error {
	r := NewHorizontal(w)
	return r.Render(root, opts)
}

func VerticalTree(w io.Writer, root *Node, opts *Options) error {
	r := NewVertical(w)
	return r.Render(root, opts)
}

func CompactTree(w io.Writer, root *Node, opts *Options) error {
	r := NewCompact(w)
	return r.Render(root, opts)
}

func SunburstTree(w io.Writer, root *Node, opts *Options) error {
	r := NewSunburst(w)
	return r.Render(root, opts)
}

func RadialTree(w io.Writer, root *Node, opts *Options) error {
	r := NewRadial(w)
	return r.Render(root, opts)
}

type Node struct {
	Value string
	Nodes []*Node
}

func NewNode(value string) *Node {
	return &Node{
		Value: value,
	}
}

func (n *Node) Leaf() bool {
	return len(n.Nodes) == 0
}

type vertical struct {
	w io.Writer
}

func NewVertical(w io.Writer) Renderer {
	return vertical{
		w: w,
	}
}

func (v vertical) Render(root *Node, options *Options) error {
	opts, err := prepareOptions(options)
	if err != nil {
		return err
	}
	opts.Orient = VerticalLayout
	var (
		layout = Ideal()
		items  = layout.Compute(root, opts)
		canvas = NewCanvas(opts.Width, opts.Height)
		screen = NewScreen(opts.Width, opts.Height)
	)

	for i := range items {
		canvas.Put(items[i].Position.X, items[i].Position.Y, items[i].Content)
		for _, x := range items[i].Children {
			paths := verticalPath(items[i], x, opts)
			for _, p := range paths {
				canvas.Connect(p)
			}
		}
	}
	if err := canvas.Render(screen); err != nil {
		return err
	}
	return screen.Render(v.w)
}

type horizontal struct {
	w io.Writer
}

func NewHorizontal(w io.Writer) Renderer {
	return horizontal{
		w: w,
	}
}

func (h horizontal) Render(root *Node, options *Options) error {
	opts, err := prepareOptions(options)
	if err != nil {
		return err
	}
	opts.Orient = HorizontalLayout
	var (
		layout = Ideal()
		items  = layout.Compute(root, opts)
		canvas = NewCanvas(opts.Width, opts.Height)
		screen = NewScreen(opts.Width, opts.Height)
	)

	for i := range items {
		canvas.Put(items[i].Position.X, items[i].Position.Y, items[i].Content)
		for _, x := range items[i].Children {
			paths := horizontalPath(items[i], x, opts)
			for _, p := range paths {
				canvas.Connect(p)
			}
		}
	}
	if err := canvas.Render(screen); err != nil {
		return err
	}
	return screen.Render(h.w)
}

type compact struct {
	w io.Writer
}

func NewCompact(w io.Writer) Renderer {
	return compact{
		w: w,
	}
}

func (c compact) Render(root *Node, options *Options) error {
	return nil
}

type treemap struct {
	w io.Writer
}

func NewTreemap(w io.Writer) Renderer {
	return treemap{
		w: w,
	}
}

func (m treemap) Render(root *Node, options *Options) error {
	return nil
}

type sunburst struct {
	w io.Writer
}

func NewSunburst(w io.Writer) Renderer {
	return sunburst{
		w: w,
	}
}

func (s sunburst) Render(root *Node, options *Options) error {
	return nil
}

type radial struct {
	w io.Writer
}

func NewRadial(w io.Writer) Renderer {
	return radial{
		w: w,
	}
}

func (r radial) Render(root *Node, options *Options) error {
	return nil
}
