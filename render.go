package trellis

import (
	"fmt"
	"io"
	"slices"
)

var renderers = map[Orientation]func(io.Writer, *Node, Options) error{
	HorizontalLayout: Horizontal,
	VerticalLayout:   Vertical,
	CompactLayout:    Compact,
}

func Render(w io.Writer, root *Node, options Options) error {
	if options == nil {
		return fmt.Errorf("options should be provided")
	}
	options.applyDefaults()
	rdr, err := options.Layout()
	if err != nil {
		return err
	}
	fn, ok := renderers[rdr.Orient]
	if !ok {
		return fmt.Errorf("unsupported layout given")
	}
	return fn(w, root, options)
}

func Horizontal(w io.Writer, root *Node, options Options) error {
	nodes, opts, err := prepareRender(root, options)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return nil
	}

	opts.estimateMinSize = estimateMinSize(root, opts.Size.Height)

	sizes, err := allocateFromSize(opts.Size, nodes, opts.AllocateMode, HorizontalLayout)
	if err != nil {
		return err
	}
	if opts.Size, err = adjustSizes(sizes, HorizontalLayout); err != nil {
		return err
	}

	master, err := NewCanvas(opts.Size)
	if err != nil {
		return err
	}
	var offset int
	for i, n := range nodes {
		clone := opts.Clone()
		clone.Size = sizes[i]

		set := stdHorizontalLayout(n, clone)
		canvas, err := NewCanvas(set.Dimension())
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
	return renderCanvas(w, master, options)
}

func Vertical(w io.Writer, root *Node, options Options) error {
	nodes, opts, err := prepareRender(root, options)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return nil
	}

	opts.estimateMinSize = estimateMinSize(root, opts.Size.Width)

	sizes, err := allocateFromSize(opts.Size, nodes, opts.AllocateMode, VerticalLayout)
	if err != nil {
		return err
	}
	if opts.Size, err = adjustSizes(sizes, VerticalLayout); err != nil {
		return err
	}

	master, err := NewCanvas(opts.Size)
	if err != nil {
		return err
	}
	var offset int
	for i, n := range nodes {
		clone := opts.Clone()
		clone.Size = sizes[i]

		set := stdVerticalLayout(n, clone)

		canvas, err := NewCanvas(set.Dimension())
		if err != nil {
			return err
		}
		for _, i := range set.Items {
			canvas.Put(i.Position.X, i.Position.Y, i.Content)
			for _, x := range i.Children {
				conn := verticalPath(i, x, clone)
				canvas.Put(conn.X(), conn.Y(), conn)
			}
		}
		canvas.Move(offset, 0)
		master.Append(canvas)
		offset += set.Width
	}
	return renderCanvas(w, master, options)
}

func Compact(w io.Writer, root *Node, options Options) error {
	if options == nil {
		return ErrOptions
	}
	opts, err := options.Layout()
	if err != nil {
		return err
	}

	opts.Spacing = SpacingL
	opts.AlignY = AlignStart
	opts.AlignX = AlignStart
	opts.Orient = CompactLayout

	set := compactLayout(root, opts)
	if ix := slices.IndexFunc(set.Items, func(i *Item) bool { return i.Root() }); ix >= 0 {
		set.Height = set.Items[ix].Weight()
	} else {
		return fmt.Errorf("missing root")
	}

	canvas, err := NewCanvas(opts.Size)
	if err != nil {
		return err
	}

	for _, i := range set.Items {
		canvas.Put(i.Position.X, i.Position.Y, i.Content)

		x := i.Position.X - compactBarWidth
		canvas.HorizontalBar(x, i.Position.Y, compactBarWidth)
		canvas.VerticalBar(x, i.Position.Y, i.Weight()+1)
	}
	return renderCanvas(w, canvas, options)
}

func Sunburst(w io.Writer, root *Node, options Options) error {
	return fmt.Errorf("not yet implemented")
}

func Radial(w io.Writer, root *Node, options Options) error {
	return fmt.Errorf("not yet implemented")
}

func TreeMap(w io.Writer, root *Node, options Options) error {
	return fmt.Errorf("not yet implemented")
}

func renderCanvas(w io.Writer, canvas *Canvas, opts Options) error {
	var (
		view View
		err  error
	)
	switch opts := opts.(type) {
	case *ScreenOptions:
		view, err = canvas.Screen(*opts)
	case *SvgOptions:
		view, err = canvas.Svg(*opts)
	case *XmlOptions:
		view, err = canvas.Xml(*opts)
	case *JsonOptions:
		view, err = canvas.Json(*opts)
	case *TableOptions:
		view, err = canvas.Table(*opts)
	default:
		return fmt.Errorf("unsupported output type")
	}
	if err != nil {
		return err
	}
	return view.Render(w)
}

func prepareRender(node *Node, options Options) ([]*Node, RenderOptions, error) {
	if options == nil {
		return nil, RenderOptions{}, fmt.Errorf("options should be provided")
	}
	opts, err := options.Layout()
	if err != nil {
		return nil, opts, err
	}

	nodes := traverse(node, opts.MinDepth)
	return nodes, opts, nil
}
