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

func traverse(node *Node, target int) []*Node {
	if target == 0 {
		return []*Node{node}
	}
	return traverseDepth(node, 0, target-1)
}

func traverseDepth(node *Node, currDepth, targetDepth int) []*Node {
	if currDepth >= targetDepth {
		return node.Nodes
	}
	var all []*Node
	for _, n := range node.Nodes {
		ns := traverseDepth(n, currDepth+1, targetDepth)
		all = append(all, ns...)
	}
	return all
}

var renderers = map[Orientation]func(io.Writer, *Node, *Options) error{
	HorizontalLayout: Horizontal,
	VerticalLayout:   Vertical,
	CompactLayout:    Compact,
}

func Render(w io.Writer, root *Node, options *Options) error {
	fn, ok := renderers[options.Orient]
	if !ok {
		return fmt.Errorf("unsupported layout given")
	}
	return fn(w, root, options)
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
	master, err := NewCanvas(opts)
	if err != nil {
		return err
	}
	var (
		// nodes = traverse(root, opts.MinDepth)
		nodes  = traverse(root, 1)
		offset int
	)
	for _, n := range nodes {
		clone := opts.Clone()
		set := stdHorizontalLayout(n, clone)

		canvas, err := NewCanvas(clone)
		if err != nil {
			return err
		}
		for _, i := range set.Items {
			canvas.Put(i.Position.X, i.Position.Y, i.Content)
			for _, x := range i.Children {
				conn := horizontalPath(i, x, clone)
				canvas.Put(conn.X(), conn.Y(), conn)
			}
		}
		canvas.Move(0, offset)
		master.Append(canvas)
		offset += set.Height
	}
	return renderCanvas(w, opts.Output, master)
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

	set := stdVerticalLayout(root, opts)
	if err := canvas.UpdateDim(set.Width, set.Height); err != nil {
		return err
	}
	for _, i := range set.Items {
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
	opts.CoordinatesStep = 0
	opts.Border = false

	set := compactLayout(root, opts)
	if ix := slices.IndexFunc(set.Items, func(i *Item) bool { return i.Root() }); ix >= 0 {
		opts.Height = set.Items[ix].Weight()
	} else {
		return fmt.Errorf("missing root")
	}

	canvas, err := NewCanvas(opts)
	if err != nil {
		return err
	}

	for _, i := range set.Items {
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
	_ = view
	return view.Render(w)
}
