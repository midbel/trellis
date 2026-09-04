package trellis

import "io"

type radial struct {
	w io.Writer
}

func NewRadial(w io.Writer) Renderer {
	return radial{
		w: w,
	}
}

func (r radial) Render(root *Node, options *Options) error {
	opts := prepareOptions(options)
	_ = opts
	return nil
}
