package trellis

import (
	"io"
)

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
	)
	adjustHorizontalWidth(layout, maker.Depth(), opts)
	adjustHorizontalHeight(maker.Spacing(), opts)

	adjustHorizontalSize(opts, maker.Depth(), maker.Spacing())

	computeHorizontalCoordinates(layout, maker.Depth(), maker.Spacing(), opts)
	if opts.Reverse {

	}
	grid := makeCanvas(opts.Width, opts.Height, opts.Border)
	// draw horizontal connectors
	drawHorizontalTree(grid, layout, opts)
	return grid.Render(h.w)
}

func adjustHorizontalSize(opts *TreeRenderOptions, depth int, spacing int) {
	var (
		sWidth  = (opts.Width / depth)
		sHeight = (opts.Height / spacing)
	)
	if w := sWidth * depth; w != opts.Width {
		opts.Width = w
	}
	if h := sHeight * spacing; h != opts.Height {
		opts.Height = h
	}
}

func adjustHorizontalWidth(layout []*layoutNode, depth int, opts *TreeRenderOptions) {
	var best int
	for _, x := range layout {
		n := len(x.Value) + (2 * (opts.Padding + opts.HorizontalGap))
		best = max(best, n)
	}
	best *= depth

	opts.Width = max(best, opts.Width)
}

func adjustHorizontalHeight(spacing int, opts *TreeRenderOptions) {
	best := spacing * (1 + opts.VerticalGap)
	opts.Height = max(best, opts.Height)
}

func computeHorizontalCoordinates(layout []*layoutNode, depth, spacing int, opts *TreeRenderOptions) {
	var (
		width   = opts.Width / depth
		height  = opts.Height / spacing
		yOffset = height / 2
		gaps    = 2 * opts.HorizontalGap
	)
	for _, x := range layout {
		var (
			value  = x.Get(opts.Padding)
			offset = getOffsetX(opts.Align, width-gaps, len(value))
		)
		x.X = (x.X * width) + offset + opts.HorizontalGap
		x.W = createSpan(x.X, x.X+width-opts.HorizontalGap)
		x.Y = (x.Y * height) + yOffset + opts.VerticalGap
		x.H = createSpan(x.Y, x.Y+height-opts.VerticalGap)
	}
}

func drawHorizontalTree(grid *canvas, layout []*layoutNode, opts *TreeRenderOptions) {
	var borderWidth int
	if opts.Border {
		borderWidth++
	}
	for _, x := range layout {
		var (
			size  = x.W.Len() + opts.HorizontalGap
			start = x.X - opts.HorizontalGap + borderWidth
		)
		if x.Leaf() {
			size = opts.HorizontalGap + x.X - x.W.Start
		} else if x.Root() {
			size -= opts.HorizontalGap
			start += opts.HorizontalGap + x.W.Offset()
		}
		grid.DrawHLine(start, x.Y+borderWidth+x.H.Offset(), size)
	}
	// // draw vertical connectors
	for _, x := range layout {
		if x.Leaf() || len(x.Children) == 1 {
			continue
		}
		var (
			first = x.Children[0]
			last  = x.Children[len(x.Children)-1]
			at    = x.X + borderWidth + x.W.Len()
		)
		grid.DrawVLine(at, first.Y+1, last.Y-first.Y+1)
	}
	// draw values
	for _, x := range layout {
		var (
			value = x.Get(opts.Padding)
			start = x.X + borderWidth
		)
		grid.Put(start, x.Y+borderWidth, value)
	}
}
