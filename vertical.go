package trellis

import (
	"io"
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
		canvas = NewCanvas(opts.Width, opts.Height)
		screen = NewScreen(opts.Width, opts.Height)
	)

	for i := range items {
		canvas.Put(items[i].X, items[i].Y, items[i].Content)
	}
	if err := canvas.Render(screen); err != nil {
		return err
	}
	return screen.Render(v.w)
}
