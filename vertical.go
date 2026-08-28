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
		maker  = makeLayout(opts.SiblingGap, opts.Anchor)
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
	best := ((opts.SiblingGap*2)+1)*depth + (2 * opts.borderWidth())
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
		z := len(x.Value) + (2 * (opts.Padding + opts.LevelGap))
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
		var (
			startX = width * n.X
			startY = height * n.Y
			value  = n.Get(opts.Padding)
			dist   = width * opts.SiblingGap
		)
		n.W = NewSpan(startX, startX+dist-opts.borderWidth())
		n.H = NewSpan(startY, startY+height-opts.borderWidth())

		n.Y = startY + opts.borderWidth()
		n.Y += n.H.Offset()

		offset := getOffsetX(opts.Align, n.W.Len(), len(value))
		if offset < 0 {
			offset = 0
		}
		n.X = startX + offset + opts.borderWidth()
	}
}

func adjustVerticalCoordinates3(layout []*layoutNode, depth, spacing int, opts *Options) {
	height := opts.Height / depth
	for _, n := range layout {
		n.Y = (height * n.Y) + opts.borderWidth()
		n.H = NewSpan(n.Y, (n.Y+height)-opts.borderWidth())
	}
	ix := slices.IndexFunc(layout, func(x *layoutNode) bool {
		return x.Root()
	})
	layout[ix].W = NewSpan(opts.borderWidth(), opts.Width+opts.borderWidth())
	computeVerticalCoordinates(layout[ix], opts)
}

func computeVerticalCoordinates(node *layoutNode, opts *Options) {
	var (
		width = node.W.Len()
		value = node.Get(opts.Padding)
	)
	node.X = node.W.Start + getOffsetX(opts.Align, node.W.Len(), len(value))
	node.Y = node.H.Start + node.H.Offset() - 1

	count := node.Weight()
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

		x.W = NewSpan(start, start+(segment*n))
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
