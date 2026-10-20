package trellis

import (
	"fmt"
	"io"
)

var renderers = map[Orientation]func([]*Item, *RenderOptions) (*Canvas, error){
	HorizontalLayout: buildHorizontal,
	VerticalLayout:   buildVertical,
	CompactLayout:    buildCompact,
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
	items, err := prepareRender(root, &rdr)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	canvas, err := fn(items, &rdr)
	if err != nil {
		return err
	}
	return renderCanvas(w, canvas, options)
}

func buildHorizontal(items []*Item, opts *RenderOptions) (*Canvas, error) {
	sizes, err := allocateFromSize(opts.Size, items, opts.AllocateMode, HorizontalLayout)
	if err != nil {
		return nil, err
	}
	if opts.Size, err = adjustSizes(sizes, HorizontalLayout); err != nil {
		return nil, err
	}
	master, err := NewCanvas(opts.Size)
	if err != nil {
		return nil, err
	}
	var offset int
	for i := range items {
		set, err := stdHorizontalLayout(items[i], cloneAndResizeOptions(*opts, sizes[i]))
		if err != nil {
			return nil, err
		}
		canvas, err := NewCanvas(set.Dimension())
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
	master, err := NewCanvas(opts.Size)
	if err != nil {
		return nil, err
	}
	var offset int
	for i := range items {
		set, err := stdVerticalLayout(items[i], cloneAndResizeOptions(*opts, sizes[i]))
		if err != nil {
			return nil, err
		}

		canvas, err := NewCanvas(set.Dimension())
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
	master, err := NewCanvas(opts.Size)
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
		canvas, err := NewCanvas(set.Dimension())
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

func cloneAndResizeOptions(opts RenderOptions, size Dimension) RenderOptions {
	clone := opts.Clone()
	clone.Size = size
	return clone
}

func prepareRender(node *Node, opts *RenderOptions) ([]*Item, error) {
	if node == nil {
		return nil, fmt.Errorf("root node should be given")
	}
	if opts.Orient == HorizontalLayout {
		opts.estimateMinSize = estimateMinSize(node, opts.Size.Height)
	} else {
		opts.estimateMinSize = estimateMinSize(node, opts.Size.Width)
	}
	return traverseNodes(traverse(node, opts.MinDepth), opts)
}

func traverseNodes(nodes []*Node, opts *RenderOptions) ([]*Item, error) {
	list := make([]*Item, 0, len(nodes))
	for _, n := range nodes {
		i, err := transformNode(n, opts)
		if err != nil {
			return nil, err
		}
		i.root = true
		list = append(list, i)
	}
	return list, nil
}

func transformNode(n *Node, opts *RenderOptions) (*Item, error) {
	it := &Item{
		Content: n.Content(),
		Metric:  opts.Metrics(n),
	}
	for _, n := range n.Nodes {
		sub, err := transformNode(n, opts)
		if err != nil {
			return nil, err
		}
		it.Children = append(it.Children, sub)
	}
	return it, nil
}
