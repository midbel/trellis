package trellis

import (
	"fmt"
	"io"
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
	items, opts, err := prepareRender(root, options)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	sizes, err := allocateFromSize(opts.Size, items, opts.AllocateMode, HorizontalLayout)
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
	for i := range items {
		clone := opts.Clone()
		clone.Size = sizes[i]

		set := stdHorizontalLayout(items[i], clone)
		canvas, err := NewCanvas(set.Dimension())
		if err != nil {
			return err
		}
		for _, i := range set.Items {
			canvas.Put(i.Position.X, i.Position.Y, i.Content)
			for _, x := range i.Children {
				path := horizontalPath(i, x, clone)
				canvas.Put(path.X(), path.Y(), path)
			}
		}
		canvas.Move(0, offset)
		master.Append(canvas)
		offset += set.Height
	}
	return renderCanvas(w, master, options)
}

func Vertical(w io.Writer, root *Node, options Options) error {
	items, opts, err := prepareRender(root, options)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	sizes, err := allocateFromSize(opts.Size, items, opts.AllocateMode, VerticalLayout)
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
	for i := range items {
		clone := opts.Clone()
		clone.Size = sizes[i]

		set := stdVerticalLayout(items[i], clone)

		canvas, err := NewCanvas(set.Dimension())
		if err != nil {
			return err
		}
		for _, i := range set.Items {
			canvas.Put(i.Position.X, i.Position.Y, i.Content)
			for _, x := range i.Children {
				path := verticalPath(i, x, clone)
				canvas.Put(path.X(), path.Y(), path)
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

	// set := compactLayout(root, opts)
	// if ix := slices.IndexFunc(set.Items, func(i *Item) bool { return i.Root() }); ix >= 0 {
	// 	set.Height = set.Items[ix].Weight()
	// } else {
	// 	return fmt.Errorf("missing root")
	// }

	// canvas, err := NewCanvas(opts.Size)
	// if err != nil {
	// 	return err
	// }

	// for _, i := range set.Items {
	// 	canvas.Put(i.Position.X, i.Position.Y, i.Content)

	// 	x := i.Position.X - compactBarWidth
	// 	canvas.HorizontalPath(x, i.Position.Y, compactBarWidth)
	// 	canvas.VerticalPath(x, i.Position.Y, i.Weight()+1)
	// }
	// return renderCanvas(w, canvas, options)
	return nil
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

func prepareRender(node *Node, options Options) ([]*Item, RenderOptions, error) {
	if options == nil {
		return nil, RenderOptions{}, fmt.Errorf("options should be provided")
	}
	opts, err := options.Layout()
	if err != nil {
		return nil, opts, err
	}
	if opts.Orient == HorizontalLayout {
		opts.estimateMinSize = estimateMinSize(node, opts.Size.Height)
	} else {
		opts.estimateMinSize = estimateMinSize(node, opts.Size.Width)
	}

	items, err := traverseNodes(traverse(node, opts.MinDepth), options)
	if err != nil {
		return nil, opts, err
	}
	// nodes := traverse(node, opts.MinDepth)
	return items, opts, nil
}

func traverseNodes(nodes []*Node, options Options) ([]*Item, error) {
	list := make([]*Item, 0, len(nodes))
	for _, n := range nodes {
		i, err := transformNode(n, options)
		if err != nil {
			return nil, err
		}
		i.root = true
		list = append(list, i)
	}
	return list, nil
}

func transformNode(n *Node, options Options) (*Item, error) {
	opts, err := options.Layout()
	if err != nil {
		return nil, err
	}
	it := &Item{
		Content: opts.Render(n, opts),
		Metric:  options.Metrics(n),
	}
	for _, n := range n.Nodes {
		sub, err := transformNode(n, options)
		if err != nil {
			return nil, err
		}
		it.Children = append(it.Children, sub)
	}
	return it, nil
}
