package trellis

type Content struct {
	Value     []byte
	Bold      bool
	Italic    bool
	Underline bool
}

type Canvas struct {
	Width  int
	Height int
	cells  []Content
}

func NewCanvas(width, height int) *Canvas {
	return &Canvas{
		Width:  width,
		Height: height,
		cells:  make([]Content, width*height),
	}
}

func (c *Canvas) Put(x, y int, content Content) error {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return nil
	}
	c.cells[y*c.Width+x] = content
	return nil
}

func (c *Canvas) Render(sc *Screen) error {
	for i, ct := range c.cells {
		y := i / c.Width
		x := i % c.Width
		sc.Put(x, y, ct)
	}
	return nil
}
