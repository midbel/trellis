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
	rdr := options.Layout()
	fn, ok := renderers[rdr.Orient]
	if !ok {
		return fmt.Errorf("unsupported layout given")
	}
	return fn(w, root, options)
}

func Horizontal(w io.Writer, root *Node, options Options) error {
	if options == nil {
		return fmt.Errorf("options should be provided")
	}

	var (
		opts   = options.Layout()
		nodes  = traverse(root, opts.MinDepth)
		offset int
	)
	master, err := NewCanvas(opts.Size)
	if err != nil {
		return err
	}
	for _, n := range nodes {
		clone := opts.Clone()
		set := stdHorizontalLayout(n, clone)

		canvas, err := NewCanvas(clone.Size)
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
	if options == nil {
		return fmt.Errorf("options should be provided")
	}

	var (
		opts = options.Layout()
		set  = stdVerticalLayout(root, opts)
	)
	canvas, err := NewCanvas(opts.Size)
	if err != nil {
		return err
	}
	if err := canvas.Resize(set.Width, set.Height); err != nil {
		return err
	}
	for _, i := range set.Items {
		canvas.Put(i.Position.X, i.Position.Y, i.Content)
		for _, x := range i.Children {
			conn := verticalPath(i, x, opts)
			canvas.Put(conn.X(), conn.Y(), conn)
		}
	}
	return renderCanvas(w, canvas, options)
}

func Compact(w io.Writer, root *Node, options Options) error {
	if options == nil {
		return ErrOptions
	}
	opts := options.Layout()

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
	// case TableOptions:
	// 	view, err = canvas.Table()
	default:
		return fmt.Errorf("unsupported output type")
	}
	if err != nil {
		return err
	}
	return view.Render(w)
}
