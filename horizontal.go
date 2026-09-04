package trellis

import (
	"io"
)

type horizontal struct {
	w io.Writer
}

func NewHorizontal(w io.Writer) Renderer {
	return horizontal{
		w: w,
	}
}

func (h horizontal) Render(root *Node, options *Options) error {
	opts := prepareOptions(options)
	opts.Orient = HorizontalLayout
	var (
		layout = Ideal()
		items  = layout.Compute(root, opts)
		canvas = NewCanvas(opts.Width, opts.Height)
		screen = NewScreen(opts.Width, opts.Height)
	)

	for i := range items {
		canvas.Put(items[i].X, items[i].Y, items[i].Content)
		for _, x := range items[i].Children {
			paths := horizontalPath(items[i], x, opts)
			for _, it := range Connector(paths) {
				canvas.Put(it.X, it.Y, it.Content)
			}
		}
	}
	if err := canvas.Render(screen); err != nil {
		return err
	}
	return screen.Render(h.w)
}
