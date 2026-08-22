package trellis

import (
	"fmt"
	"io"
	"slices"
)

type vertical struct {
	w io.Writer
}

func NewVertical(w io.Writer) Renderer {
	return vertical{
		w: w,
	}
}

func (v vertical) Render(tree *Tree, options *Options) error {
	var (
		opts   = prepareOptions(options)
		maker  = makeLayout(opts.VerticalGap, opts.Position)
		layout = maker.Make(tree.Root)
	)
	adjustVerticalWidth(layout, opts)
	adjustVerticalHeight(layout, maker.Depth(), opts)

	adjustVerticalSize(opts, maker.Depth(), maker.Spacing())
	for i := range layout {
		layout[i].X, layout[i].Y = layout[i].Y, layout[i].X
	}
	if opts.Reverse {
		reverseVerticalCoordinates(maker, layout)
	}
	adjustVerticalCoordinates(layout, maker.Depth(), maker.Spacing(), opts)
	grid := makeCanvas(opts.Width, opts.Height, opts.Border)
	ix := slices.IndexFunc(layout, func(x *layoutNode) bool {
		return x.Root()
	})
	drawVerticalTree(grid, layout[ix], opts)
	return grid.Render(v.w)
}

func adjustVerticalSize(opts *Options, depth, spacing int) {
	sHeight := (opts.Height / depth)
	if h := sHeight * depth; h != opts.Height {
		opts.Height = h
	}
}

func reverseVerticalCoordinates(maker *layoutMaker, layout []*layoutNode) {
	depth := maker.Depth()
	for _, x := range layout {
		x.Y = depth - x.Y - 1
	}
}

func adjustVerticalHeight(layout []*layoutNode, depth int, opts *Options) {
	best := ((opts.VerticalGap*2)+1)*depth + (2 * opts.borderWidth())
	opts.Height = max(best, opts.Height)
}

func adjustVerticalWidth(layout []*layoutNode, opts *Options) {
	var (
		sizes  = make(map[int]int)
		border int
		best   int
	)
	if opts.Border {
		border++
	}
	for _, x := range layout {
		z := len(x.Value) + (2 * (opts.Padding + opts.HorizontalGap))
		sizes[x.X] += z
	}
	for _, z := range sizes {
		best = max(best, z+(2*border))
	}
	opts.Width = max(best, opts.Width)
}

func adjustVerticalCoordinates(layout []*layoutNode, depth, spacing int, opts *Options) {
	var (
		width  = opts.Width / spacing
		height = opts.Height / depth
	)
	for _, n := range layout {
		n.X = width * n.X
		n.Y = (height * n.Y) + opts.borderWidth()
		n.H = createSpan(n.Y, (n.Y+height)-opts.borderWidth())
	}
	ix := slices.IndexFunc(layout, func(x *layoutNode) bool {
		return x.Root()
	})
	layout[ix].W = createSpan(opts.borderWidth(), opts.Width+opts.borderWidth())
	computeVerticalCoordinates(layout[ix], opts)
}

func computeVerticalCoordinates(node *layoutNode, opts *Options) {
	var (
		width = node.W.Len()
		value = node.Get(opts.Padding)
	)
	node.X = node.W.Start + getOffsetX(opts.Align, node.W.Len(), len(value))
	node.Y = node.H.Start + node.H.Offset() - 1

	var count int
	for _, x := range node.Children {
		count += x.Weight()
	}
	if count == 0 {
		count++
	}
	var (
		segment = width / count
		rem     = width % count
		start   = node.W.Start
	)
	for _, x := range node.Children {
		n := x.Weight()

		x.W = createSpan(start, start+(segment*n))
		if len(node.Children) == 1 {
			x.W = node.W
		}
		var addOne bool
		if rem > 0 && len(x.Children) > 1 {
			rem--
			x.W.End++
			addOne = true
		}
		computeVerticalCoordinates(x, opts)
		start += (segment * n)
		if addOne {
			start++
		}
	}
}

func drawVerticalTree(grid *canvas, node *layoutNode, opts *Options) {
	fmt.Println(">>", node.Value, node.X, node.Y)
	grid.Put(node.X, node.Y, node.Get(opts.Padding))
	for _, x := range node.Children {
		drawVerticalTree(grid, x, opts)

		// draw connectors
		var (
			source = node.Anchor(opts.Padding)
			target = x.Anchor(opts.Padding)
		)
		if target == source {
			var (
				start = node.Y
				dist  = x.Y - node.Y
			)
			if opts.Reverse {
				start = x.Y
				dist = node.Y - x.Y
			}
			grid.DrawVLine(source, start+opts.borderWidth(), dist-opts.borderWidth())
		} else {
			if opts.Reverse {
				grid.DrawVLine(target, x.Y+opts.borderWidth(), node.H.Start-x.Y)
				grid.DrawVLine(source, node.H.Start, node.Y-node.H.Start)
			} else {
				grid.DrawVLine(source, node.Y+opts.borderWidth(), x.H.Start-node.Y)
				grid.DrawVLine(target, x.H.Start, x.Y-x.H.Start)
			}

			var (
				start  = x.H.Start
				anchor = source
				dist   = target - source
			)
			if opts.Reverse {
				start = node.H.Start
			}
			if dist < 0 {
				anchor = target
				dist = -dist
			}
			grid.DrawHLine(anchor, start, dist)
		}
	}
}
