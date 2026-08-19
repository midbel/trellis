package trellis

import (
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
	var border int
	if opts.Border {
		border++
	}
	best := ((opts.VerticalGap*2)+1)*depth + (2 * border)
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
		width       = opts.Width / spacing
		height      = opts.Height / depth
		borderWidth int
	)
	for _, n := range layout {
		n.X = width * n.X
		n.Y = height * n.Y
	}
	if opts.Border {
		borderWidth++
	}
	ix := slices.IndexFunc(layout, func(x *layoutNode) bool {
		return x.Root()
	})

	layout[ix].W = createSpan(borderWidth, opts.Width+borderWidth)
	layout[ix].H = createSpan(borderWidth, height+borderWidth)
	computeVerticalCoordinates(layout[ix], opts)
}

func computeVerticalCoordinates(node *layoutNode, opts *Options) {
	var count int
	for _, x := range node.Children {
		count += x.Weight()
	}
	if count == 0 {
		count++
	}
	var (
		width = node.W.Len() / count
		rem   = node.W.Len() % count
		start = node.W.Start
	)
	for _, x := range node.Children {
		n := x.Weight()

		x.W = createSpan(start, start+(width*n))
		if len(node.Children) == 1 {
			x.W = node.W
		}
		var addOne bool
		if rem > 0 && len(x.Children) > 1 {
			rem--
			x.W.End++
			addOne = true
		}
		x.H = node.H.Next()
		computeVerticalCoordinates(x, opts)
		start += (width * n)
		if addOne {
			start++
		}
	}
}

func drawVerticalTree(grid *canvas, node *layoutNode, opts *Options) {
	y := node.Y + node.H.Offset()
	grid.Put(node.W.CenterValue(node.Value, opts.Padding), y, node.Get(opts.Padding))

	var (
		spans []span
		y1    = node.Y + node.H.Len()
		dist  = y1 - (y + 1)
	)
	for _, x := range node.Children {
		if x.Leaf() {
			if opts.Reverse {
				grid.DrawVLine(x.W.Center(), node.Y-dist, dist+1)
			} else {
				grid.DrawVLine(x.W.Center(), y1, 1+node.H.Offset())
			}
		} else {
			if opts.Reverse {
				grid.DrawVLine(x.W.Center(), node.Y-dist, dist+1)
			} else {
				grid.DrawVLine(x.W.Center(), y1, dist+1)
			}
		}

		spans = append(spans, x.W)
		drawVerticalTree(grid, x, opts)
	}
	if len(spans) >= 2 {
		first, last := spans[0], spans[len(spans)-1]
		if opts.Reverse {
			grid.DrawHLine(first.Center(), node.Y, last.Center()-first.Center())
		} else {
			grid.DrawHLine(first.Center(), node.Y+node.H.Len(), last.Center()-first.Center())
		}
	}
	if !node.Leaf() {
		if opts.Reverse {
			grid.DrawVLine(node.W.Center(), node.Y+1, node.H.Offset()-1)
		} else {
			grid.DrawVLine(node.W.Center(), y+1, dist)
		}
	}
}
