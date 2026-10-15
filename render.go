package trellis

import (
	"fmt"
	"io"
)

var renderers = map[Orientation]func(io.Writer, *Node, Options) error{
	HorizontalLayout: horizontalTree,
	VerticalLayout:   verticalTree,
	CompactLayout:    compactTree,
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

func horizontalTree(w io.Writer, root *Node, options Options) error {
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
			if err := canvas.Put(i.X(), i.Y(), i.Content); err != nil {
				return err
			}
			for _, x := range i.Children {
				path := horizontalPath(i, x)
				if err := canvas.Put(path.X(), path.Y(), path); err != nil {
					return err
				}
			}
		}
		canvas.Move(0, offset)
		master.Append(canvas)
		offset += set.Height
	}
	return renderCanvas(w, master, options)
}

func verticalTree(w io.Writer, root *Node, options Options) error {
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
			if err := canvas.Put(i.X(), i.Y(), i.Content); err != nil {
				return err
			}
			for _, x := range i.Children {
				path := verticalPath(i, x)
				if err := canvas.Put(path.X(), path.Y(), path); err != nil {
					return err
				}
			}
		}
		canvas.Move(offset, 0)
		master.Append(canvas)
		offset += set.Width
	}
	return renderCanvas(w, master, options)
}

func compactTree(w io.Writer, root *Node, options Options) error {
	items, opts, err := prepareRender(root, options)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	opts.Placement = AlignStart
	master, err := NewCanvas(opts.Size)
	if err != nil {
		return err
	}
	var offset int
	for i := range items {
		var (
			clone = opts.Clone()
			set   = compactLayout(items[i], clone)
		)
		canvas, err := NewCanvas(set.Dimension())
		if err != nil {
			return err
		}
		for _, i := range set.Items {
			if err := canvas.Put(i.Position.X, i.Position.Y, i.Content); err != nil {
				return err
			}
			p1 := NewPoint(0, i.Position.Y)
			p2 := NewPoint(i.Position.X-1, i.Position.Y)
			p := NewPath(p1, p2)
			if p.DistanceX() > 0 {
				canvas.Put(p.X(), p.Y(), p)
			}
		}
		canvas.Move(0, offset)
		master.Append(canvas)
		offset += set.Height
	}
	return renderCanvas(w, master, options)
}

func sunburst(w io.Writer, root *Node, options Options) error {
	return fmt.Errorf("not yet implemented")
}

func radial(w io.Writer, root *Node, options Options) error {
	return fmt.Errorf("not yet implemented")
}

func treeMap(w io.Writer, root *Node, options Options) error {
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
