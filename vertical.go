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

func (v vertical) Render(tree *Tree, opts *TreeRenderOptions) error {
	if opts == nil {
		opts = defaultTreeRenderOptions.clone()
	} else {
		opts = opts.clone()
	}
	var (
		maker  = makeLayout(opts.VerticalGap, opts.Position)
		layout = maker.Make(tree.Root)
		bWidth int
	)
	if opts.Border {
		bWidth++
	}
	adjustVerticalWidth(layout, bWidth, opts)
	adjustVerticalHeight(layout, maker.Depth(), bWidth, opts)

	var (
		sWidth  = (opts.Width / maker.Spacing())
		sHeight = (opts.Height / maker.Depth())
	)
	if h := sHeight * maker.Depth(); h != opts.Height {
		opts.Height = h
	}

	for i := range layout {
		layout[i].X, layout[i].Y = layout[i].Y, layout[i].X
	}
	for _, n := range layout {
		n.X = sWidth * n.X
		n.Y = sHeight * n.Y
	}
	grid := makeCanvas(opts.Width, opts.Height, opts.Border)
	// draw values
	ix := slices.IndexFunc(layout, func(x *layoutNode) bool {
		return x.Root()
	})

	layout[ix].W = createSpan(bWidth, opts.Width+bWidth)
	layout[ix].H = createSpan(bWidth, bWidth+sHeight)
	computeVerticalCoordinates(layout[ix], opts)
	if opts.Reverse {

	}
	drawVerticalTree(grid, layout[ix], opts)
	return grid.Render(v.w)
}

func adjustVerticalHeight(layout []*layoutNode, depth, border int, opts *TreeRenderOptions) {
	best := ((opts.VerticalGap*2)+1)*depth + (2 * border)
	opts.Height = max(best, opts.Height)
}

func adjustVerticalWidth(layout []*layoutNode, border int, opts *TreeRenderOptions) {
	var (
		sizes = make(map[int]int)
		best  int
	)
	for _, x := range layout {
		z := len(x.Value) + (2 * (opts.Padding + opts.HorizontalGap))
		sizes[x.X] += z
	}
	for _, z := range sizes {
		best = max(best, z+(2*border))
	}
	opts.Width = max(best, opts.Width)
}

func computeVerticalCoordinates(node *layoutNode, opts *TreeRenderOptions) {
	var count int
	for _, x := range node.Children {
		count += x.Weight()
	}
	if count == 0 {
		count++
	}
	var (
		width  = node.W.Len() / count
		rem    = node.W.Len() % count
		offset = node.W.Start
	)
	for _, x := range node.Children {
		n := x.Weight()

		x.W = createSpan(offset, offset+(width*n))
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
		offset += (width * n)
		if addOne {
			offset++
		}
	}
}

func drawVerticalTree(grid *canvas, node *layoutNode, opts *TreeRenderOptions) {
	y := node.Y + node.H.Offset()
	grid.Put(node.W.CenterValue(node.Value, opts.Padding), y, node.Get(opts.Padding))

	var (
		spans []span
		y1    = node.Y + node.H.Len()
		dist  = y1 - (y + 1)
	)
	for _, x := range node.Children {
		if x.Leaf() {
			grid.DrawVLine(x.W.Center(), y1, 1+node.H.Offset())
		} else {
			grid.DrawVLine(x.W.Center(), y1, dist+1)
		}

		spans = append(spans, x.W)
		drawVerticalTree(grid, x, opts)
	}
	if len(spans) >= 2 {
		first, last := spans[0], spans[len(spans)-1]
		grid.DrawHLine(first.Center(), y1, last.Center()-first.Center())
	}
	if !node.Leaf() {
		grid.DrawVLine(node.W.Center(), y+1, dist)
	}
}
