package trellis

import "io"

type horizontal struct {
	w io.Writer
}

func NewHorizontal(w io.Writer) Renderer {
	return horizontal{
		w: w,
	}
}

func (h horizontal) Render(tree *Tree, opts *TreeRenderOptions) error {
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
	adjustHorizontalWidth(layout, maker.Depth(), opts)
	adjustHorizontalHeight(maker.Spacing(), opts)

	var (
		sWidth  = (opts.Width / maker.Depth())
		sHeight = (opts.Height / maker.Spacing())
		vOffset = sHeight / 2
	)
	if w := sWidth * maker.Depth(); w != opts.Width {
		opts.Width = w
	}
	if h := sHeight * maker.Spacing(); h != opts.Height {
		opts.Height = h
	}
	computeHorizontalCoordinates(layout, maker.Depth(), maker.Spacing(), opts)

	grid := makeCanvas(opts.Width, opts.Height, opts.Border)
	// draw horizontal connectors
	for _, x := range layout {
		var (
			size   = sWidth
			offset = getOffsetX(opts.Align, sWidth-(2*opts.HorizontalGap), len(x.Value)+(2*opts.Padding))
			start  = x.X - opts.HorizontalGap + bWidth
		)
		if x.Leaf() {
			size = opts.HorizontalGap + offset
		} else if x.Root() {
			size -= opts.HorizontalGap
			start += opts.HorizontalGap + offset
		}
		grid.DrawHLine(start, x.Y+bWidth+vOffset, size)
	}
	// draw vertical connectors
	for _, x := range layout {
		if x.Leaf() {
			continue
		}
		for i := 0; i < len(x.Children)-1; i++ {
			var (
				start  = x.Children[i].Y + vOffset + bWidth
				length = x.Children[i+1].Y + vOffset + bWidth - start
				at     = x.X + bWidth + sWidth - opts.HorizontalGap
			)
			grid.DrawVLine(at, start, length)
		}
	}
	// draw values
	for _, x := range layout {
		var (
			value  = x.Get(opts.Padding)
			size   = len(value)
			offset = getOffsetX(opts.Align, sWidth-(2*opts.HorizontalGap), size)
			start  = x.X + bWidth + offset
		)
		grid.Put(start, x.Y+bWidth+vOffset, value)
	}
	return grid.Render(h.w)
}

func adjustHorizontalWidth(layout []*layoutNode, depth int, opts *TreeRenderOptions) {
	var best int
	if opts.Width == 0 {
		for _, x := range layout {
			n := len(x.Value) + (2 * (opts.Padding + opts.HorizontalGap))
			best = max(best, n)
		}
		best *= depth
	}
	opts.Width = max(best, opts.Width)
}

func adjustHorizontalHeight(spacing int, opts *TreeRenderOptions) {
	best := spacing * opts.VerticalGap
	opts.Height = max(best, opts.Height)
}

func computeHorizontalCoordinates(layout []*layoutNode, depth, spacing int, opts *TreeRenderOptions) {
	for _, x := range layout {
		x.X = ((x.X * opts.Width) / depth) + opts.HorizontalGap
		x.Y = (x.Y * opts.Height) / spacing
	}
}

func drawHorizontalTree(grid *canvas, layout []*layoutNode, opts *TreeRenderOptions) {

}
