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

func (h horizontal) Render(tree *Tree, options *Options) error {
	var (
		opts   = prepareOptions(options)
		maker  = makeLayout(opts.VerticalGap, opts.Position)
		layout = maker.Make(tree.Root)
	)
	adjustHorizontalWidth(layout, maker.Depth(), opts)
	adjustHorizontalHeight(maker.Spacing(), opts)

	adjustHorizontalSize(opts, maker.Depth(), maker.Spacing())

	if opts.Reverse {
		reverseHorizontalCoordinates(maker, layout)
	}

	computeHorizontalCoordinates(layout, maker.Depth(), maker.Spacing(), opts)
	grid := makeCanvas(opts.Width, opts.Height, opts.Border)
	drawHorizontalTree(grid, layout, opts)
	return grid.Render(h.w)
}

func reverseHorizontalCoordinates(maker *layoutMaker, layout []*layoutNode) {
	depth := maker.Depth()
	for _, x := range layout {
		x.X = depth - x.X - 1
	}
}

func adjustHorizontalSize(opts *Options, depth, spacing int) {
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

func adjustHorizontalWidth(layout []*layoutNode, depth int, opts *Options) {
	var best int
	for _, x := range layout {
		n := len(x.Value) + (2 * (opts.Padding + opts.HorizontalGap))
		best = max(best, n)
	}
	best *= depth

	opts.Width = max(best, opts.Width)
}

func adjustHorizontalHeight(spacing int, opts *Options) {
	best := spacing * (1 + opts.VerticalGap)
	opts.Height = max(best, opts.Height)
}

func computeHorizontalCoordinates(layout []*layoutNode, depth, spacing int, opts *Options) {
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
			startX = (x.X * width) + opts.HorizontalGap
			startY = x.Y * height
		)
		x.X = startX + offset + opts.HorizontalGap
		x.W = createSpan(startX, startX+width-gaps)
		x.Y = startY + yOffset + opts.VerticalGap
		x.H = createSpan(startY, startY+height-opts.VerticalGap)
	}
}

func drawHorizontalTree(grid *canvas, layout []*layoutNode, opts *Options) {
	var borderWidth int
	if opts.Border {
		borderWidth++
	}
	// draw horizontal connectors
	for _, x := range layout {
		var (
			size  = x.W.Len()
			start = x.W.Start + borderWidth
		)
		if x.Leaf() {
			if opts.Reverse {
				size = x.W.End - x.X - opts.Padding + opts.HorizontalGap + 1
				start = x.X + opts.Padding
			} else {
				size = x.X - x.W.Start + opts.HorizontalGap
				start -= opts.HorizontalGap
			}
		} else if x.Root() {
			if !opts.Reverse {
				size = x.W.End - x.X - opts.Padding + opts.HorizontalGap + 1
				start = x.X + opts.Padding
			} else {
				size = x.X - x.W.Start + opts.HorizontalGap
				start -= opts.HorizontalGap
			}
		} else {
			start -= opts.HorizontalGap
			size += 2 * opts.HorizontalGap
		}
		grid.DrawHLine(start, x.Y+borderWidth+x.H.Offset(), size)
	}
	// draw vertical connectors
	for _, x := range layout {
		if x.Leaf() || len(x.Children) == 1 {
			continue
		}
		var (
			first = x.Children[0]
			last  = x.Children[len(x.Children)-1]
			at    = x.W.Start + borderWidth
		)
		if !opts.Reverse {
			at += x.W.Len() + opts.HorizontalGap
		} else {
			at -= opts.HorizontalGap
		}
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
