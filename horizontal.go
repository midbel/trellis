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
		maker  = makeLayout(opts.SiblingGap, opts.Anchor)
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
		n := len(x.Value) + (2 * (opts.Padding + opts.LevelGap))
		best = max(best, n)
	}
	best *= depth

	opts.Width = max(best, opts.Width)
}

func adjustHorizontalHeight(spacing int, opts *Options) {
	best := spacing * (1 + opts.SiblingGap)
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
			startX = x.X * width
			startY = x.Y * height
		)
		x.W = createSpan(startX+opts.LevelGap, startX+width-opts.LevelGap)
		x.X = x.W.Start + getOffsetX(opts.Align, x.W.Len(), len(value))
		x.H = createSpan(startY, startY+height-opts.SiblingGap)
		x.Y = x.H.Start + yOffset
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
		if source == target {
			dist := x.X - start - opts.borderWidth()
			if opts.Reverse {
				dist = node.X - start - opts.borderWidth()
			}
			grid.DrawHLine(start, source+opts.borderWidth(), dist)
		} else {
			dist := node.W.End - start + opts.LevelGap
			if opts.Reverse {
				dist = x.W.End - start + opts.LevelGap
			}
			grid.DrawHLine(start, source+opts.borderWidth(), dist)

			dist = x.X - x.W.Start + opts.LevelGap - opts.borderWidth()
			start = x.W.Start - opts.LevelGap
			if opts.Reverse {
				start = node.W.Start - opts.LevelGap
				dist = node.X - node.W.Start + opts.LevelGap - opts.borderWidth()
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
			grid.DrawVLine(start-opts.LevelGap, anchor+opts.borderWidth(), dist)
		}
	}
	grid.Put(node.X, node.Y+opts.borderWidth(), label)
}
