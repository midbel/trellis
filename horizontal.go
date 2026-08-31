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
	opts := prepareOptions(options)
	opts.Orient = HorizontalLayout
	var (
		layout = Ideal()
		items  = layout.Compute(tree, opts)
	)

	ix := slices.IndexFunc(items, func(n *Item) bool {
		return n.Root()
	})
	_ = ix
	return nil
}
