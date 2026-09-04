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

func (s sunburst) Render(root *Node, options *Options) error {
	return nil
}
