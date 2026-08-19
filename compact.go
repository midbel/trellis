package trellis

import "io"

type compact struct {
	w io.Writer
}

func NewCompact(w io.Writer) Renderer {
	return compact{
		w: w,
	}
}

func (c compact) Render(tree *Tree, options *Options) error {
	opts := prepareOptions(options)
	_ = opts
	return nil
}
