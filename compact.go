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

func (c compact) Render(tree *Tree, opts *TreeRenderOptions) error {
	if opts == nil {
		opts = defaultTreeRenderOptions.clone()
	} else {
		opts = opts.clone()
	}
	return nil
}
