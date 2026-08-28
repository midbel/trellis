package trellis

import (
	"bufio"
	"io"
)

type canvas struct {
	Width  int
	Height int
	cells  [][]byte
}

func makeCanvas(width, height int, border bool) *canvas {
	if border {
		width += 2
		height += 2
	}

	var (
		cells = make([][]byte, height)
		blank = make([]byte, width)
	)
	for i := range blank {
		blank[i] = ' '
	}
	if border {
		blank[0] = verticalBarAscii
		blank[width-1] = verticalBarAscii
	}
	for i := range cells {
		cells[i] = make([]byte, width)
		if border && (i == 0 || i == height-1) {
			for j := range cells[i] {
				cells[i][j] = horizontalBarAscii
				if j == 0 || j == width-1 {
					cells[i][j] = connectBarAscii
				}
			}
		} else {
			copy(cells[i], blank)
		}
	}
	c := &canvas{
		cells:  cells,
		Width:  width,
		Height: height,
	}
	return c
}

func (c *canvas) DrawHLine(x, y, length int) {
	tmp := make([]byte, length+1)
	for i := range tmp {
		if tmp[i] == connectBarAscii {
			continue
		}
		tmp[i] = horizontalBarAscii
		if i == 0 || i == length {
			tmp[i] = connectBarAscii
		}
	}
	c.Put(x, y, tmp)
}

func (c *canvas) DrawVLine(x, y, length int) {
	for i := 0; i < length; i++ {
		at := y + i
		if b := c.GetByte(x, at); b == connectBarAscii {
			continue
		} else if b == horizontalBarAscii {
			c.PutByte(x, at, connectBarAscii)
		} else {
			c.PutByte(x, at, verticalBarAscii)
		}
	}
}

func (c *canvas) GetByte(x, y int) byte {
	return c.cells[y][x]
}

func (c *canvas) PutByte(x, y int, char byte) {
	c.cells[y][x] = char
}

func (c *canvas) Put(x, y int, chars []byte) {
	copy(c.cells[y][x:], chars)
}

func (c *canvas) Render(w io.Writer) error {
	ws := bufio.NewWriter(w)
	defer ws.Flush()
	for i := range c.cells {
		_, err := ws.Write(c.cells[i])
		if err != nil {
			return err
		}
		ws.WriteByte('\n')
	}
	return nil
}

func drawVerticalTree(grid *canvas, node *layoutNode, opts *Options) {
	grid.Put(node.X, node.Y, node.Get(opts.Padding))
	for _, x := range node.Children {
		drawVerticalTree(grid, x, opts)

		// draw connectors
		var (
			source = node.Anchor(opts.Padding)
			target = x.Anchor(opts.Padding)
		)
		if target == source {
			var (
				start = node.Y
				dist  = x.Y - node.Y
			)
			if opts.Reverse {
				start = x.Y
				dist = node.Y - x.Y
			}
			grid.DrawVLine(source, start+opts.borderWidth(), dist-opts.borderWidth())
		} else {
			if opts.Reverse {
				grid.DrawVLine(target, x.Y+opts.borderWidth(), node.H.Start-x.Y)
				grid.DrawVLine(source, node.H.Start, node.Y-node.H.Start)
			} else {
				grid.DrawVLine(source, node.Y+opts.borderWidth(), x.H.Start-node.Y)
				grid.DrawVLine(target, x.H.Start, x.Y-x.H.Start)
			}

			var (
				start  = x.H.Start
				anchor = source
				dist   = target - source
			)
			if opts.Reverse {
				start = node.H.Start
			}
			if dist < 0 {
				anchor = target
				dist = -dist
			}
			grid.DrawHLine(anchor, start, dist)
		}
	}
}

func drawHorizontalTree(grid *canvas, node *Item, opts *Options) {
	start := node.X + len(node.Value)
	for _, x := range node.Children {
		var (
			source = x.Y
			target = node.Y
		)
		if source == target {
			if opts.Reverse {
				start = x.X + len(x.Value)
				grid.DrawHLine(start, source, node.W.Start-start-1)
			} else {
				grid.DrawHLine(start, source, x.W.Start-start-1)
			}
		} else {
			var (
				distance = x.W.Start - start
				left     int
				right    int
			)
			if mid := distance / 2; mid+mid == distance {
				left, right = mid, mid
			} else {
				left, right = mid+1, mid
			}
			grid.DrawHLine(start, node.Y, left)
			grid.DrawHLine(start+left, x.Y, right-1)

			anchor := source
			distance = target - source
			if distance < 0 {
				distance = -distance
				anchor = target
			}
			grid.DrawVLine(start+left, anchor, distance)
		}
		drawHorizontalTree(grid, x, opts)
	}
	grid.Put(node.X, node.Y, node.Value)
}
