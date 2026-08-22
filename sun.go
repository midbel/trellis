package trellis

import "io"

type sunburst struct {
	w io.Writer
}

func NewSunburst(w io.Writer) Renderer {
	return sunburst{
		w: w,
	}
}

func (s sunburst) Render(tree *Tree, options *Options) error {
	return nil
}
