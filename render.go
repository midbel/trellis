package trellis

import (
	"fmt"
	"io"
	"slices"
)

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

func Table(w io.Writer, root *Node, options *Options) error {
	return nil
}

func Horizontal(w io.Writer, root *Node, options *Options) error {
	opts, err := prepareOptions(options)
	if err != nil {
		return err
	}
	opts.Orient = HorizontalLayout
	canvas, err := NewCanvas(opts)
	if err != nil {
		return err
	}

	items := stdHorizontalLayout(root, opts)
	for _, i := range items {
		canvas.Put(i.Position.X, i.Position.Y, i.Content)
		for _, x := range i.Children {
			conn := horizontalPath(i, x, opts)
			canvas.Put(conn.X(), conn.Y(), conn)
		}
	}
	return renderCanvas(w, opts.Output, canvas)
}

func Vertical(w io.Writer, root *Node, options *Options) error {
	opts, err := prepareOptions(options)
	if err != nil {
		return err
	}
	opts.Orient = VerticalLayout

	canvas, err := NewCanvas(opts)
	if err != nil {
		return err
	}

	items := stdVerticalLayout(root, opts)
	for _, i := range items {
		canvas.Put(i.Position.X, i.Position.Y, i.Content)
		for _, x := range i.Children {
			conn := verticalPath(i, x, opts)
			canvas.Put(conn.X(), conn.Y(), conn)
		}
	}
	return renderCanvas(w, opts.Output, canvas)
}

func Compact(w io.Writer, root *Node, options *Options) error {
	opts, err := prepareOptions(options)
	if err != nil {
		return err
	}
	opts.Spacing = SpacingL
	opts.AlignY = AlignStart
	opts.AlignX = AlignStart
	opts.Orient = CompactLayout

	items := compactLayout(root, opts)
	if ix := slices.IndexFunc(items, func(i *Item) bool { return i.Root() }); ix >= 0 {
		opts.Height = items[ix].Weight()
	} else {
		return fmt.Errorf("missing root")
	}

	canvas, err := NewCanvas(opts)
	if err != nil {
		return err
	}

	for _, i := range items {
		canvas.Put(i.Position.X, i.Position.Y, i.Content)

		x := i.Position.X - compactBarWidth
		canvas.HorizontalBar(x, i.Position.Y, compactBarWidth)
		canvas.VerticalBar(x, i.Position.Y, i.Weight()+1)
	}
	return renderCanvas(w, opts.Output, canvas)
}

func Sunburst(w io.Writer, root *Node, options *Options) error {
	return fmt.Errorf("not yet implemented")
}

func Radial(w io.Writer, root *Node, options *Options) error {
	return fmt.Errorf("not yet implemented")
}

func TreeMap(w io.Writer, root *Node, options *Options) error {
	return fmt.Errorf("not yet implemented")
}

func renderCanvas(w io.Writer, out Output, canvas *Canvas) error {
	var (
		view View
		err  error
	)
	switch out {
	case OutputScreen:
		view, err = canvas.Screen()
	case OutputSvg:
		view, err = canvas.Svg()
	case OutputXml:
		view, err = canvas.Xml()
	case OutputJson:
		view, err = canvas.Json()
	default:
		return fmt.Errorf("no output provided")
	}
	if err != nil {
		return err
	}
	return view.Render(w)
}
