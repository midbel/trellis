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
	opts := prepareOptions(options)
	opts.Orient = VerticalLayout
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
