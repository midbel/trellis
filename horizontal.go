package trellis

import (
	"io"
	"slices"
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

	ix := slices.IndexFunc(layout, func(n *layoutNode) bool {
		return n.Root()
	})
	drawHorizontalTree(grid, layout[ix], opts)
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
	)
	for _, x := range layout {
		var (
			value  = x.Get(opts.Padding)
			offset = getOffsetX(opts.Align, width-opts.hGaps(), len(value))
			startX = (x.X * width) + opts.HorizontalGap
			startY = x.Y * height
		)
		x.X = startX + offset + opts.HorizontalGap
		x.W = createSpan(startX, startX+width-opts.hGaps())
		x.Y = startY + yOffset + opts.VerticalGap
		x.H = createSpan(startY, startY+height-opts.VerticalGap)
	}
}

func drawHorizontalTree(grid *canvas, node *layoutNode, opts *Options) {
	var (
		label = opts.paddedValue(node.Value)
		size  = len(label)
	)
	for _, x := range node.Children {
		drawHorizontalTree(grid, x, opts)

		var (
			source = node.Y
			target = x.Y
			start  = node.X + size
			value  = label
		)
		if opts.Reverse {
			value = opts.paddedValue(x.Value)
			size = len(value)
			start = x.X + size

			source, target = target, source
		}
		_ = start
		if source == target {
			dist := x.X - start - opts.borderWidth()
			if opts.Reverse {
				dist = node.X - start - opts.borderWidth()
			}
			grid.DrawHLine(start, source+opts.borderWidth(), dist)
		} else {
			dist := node.W.End - start + opts.HorizontalGap
			if opts.Reverse {
				dist = x.W.End - start + opts.HorizontalGap
			}
			grid.DrawHLine(start, source+opts.borderWidth(), dist)

			dist = x.X - x.W.Start + opts.HorizontalGap - opts.borderWidth()
			start = x.W.Start - opts.HorizontalGap
			if opts.Reverse {
				start = node.W.Start - opts.HorizontalGap
				dist = node.X - node.W.Start + opts.HorizontalGap - opts.borderWidth()
			}
			grid.DrawHLine(start, target+opts.borderWidth(), dist)

			anchor := source
			dist = target - source
			if dist < 0 {
				dist = -dist
				anchor = target
			}
			start = x.W.Start
			if opts.Reverse {
				start = node.W.Start
			}
			grid.DrawVLine(start-opts.HorizontalGap, anchor+opts.borderWidth(), dist)
		}
	}
	grid.Put(node.X, node.Y+opts.borderWidth(), label)
}
