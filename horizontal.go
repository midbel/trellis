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
	if opts.Width == 0 {
		for _, x := range layout {
			n := len(x.Value) + (2 * (opts.Padding + opts.HorizontalGap))
			opts.Width = max(opts.Width, n)
		}
		opts.Width = opts.Width * maker.Depth()
	} else {
		opts.Width += (2 * opts.HorizontalGap) * maker.Depth()
	}
	if opts.Height == 0 {
		opts.Height = maker.Spacing() * opts.VerticalGap
	}

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
	for _, x := range layout {
		x.X = ((x.X * opts.Width) / maker.Depth()) + opts.HorizontalGap
		x.Y = (x.Y * opts.Height) / maker.Spacing()
	}

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
