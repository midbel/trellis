package trellis

import (
	"bufio"
	"io"
)

type compact struct {
	w *bufio.Writer
}

func NewCompact(w io.Writer) Renderer {
	return compact{
		w: bufio.NewWriter(w),
	}
}

func (c compact) Render(tree *Tree, options *Options) error {
	return nil
}
