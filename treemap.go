package trellis

import "io"

type treemap struct {
	w io.Writer
}

func NewTreemap(w io.Writer) Renderer {
	return treemap{
		w: w,
	}
}

func (m treemap) Render(tree *Tree, options *Options) error {
	opts := prepareOptions(options)
	_ = opts
	return nil
}
