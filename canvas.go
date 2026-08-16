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
		if c.GetByte(x, at) == connectBarAscii {
			continue
		}
		c.PutByte(x, at, verticalBarAscii)
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
