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

func (c *Canvas) At(x, y int) Content {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return Content{}
	}
	return c.cells[y*c.Width+x]
}

func (c *Canvas) Put(x, y int, content Content) {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return
	}
	c.cells[y*c.Width+x] = content
}
