package trellis

import (
	"bufio"
	"fmt"
	"io"
)

const space = ' '

type XmlFile struct {
	cells []Cell
}

func (f *XmlFile) Put(x, y int, char rune) error {
	return nil
}

func (f *XmlFile) Render(w io.Writer) error {
	return nil
}

type Screen struct {
	lines  [][]rune
	dim    Dimension
}

func NewScreen(width, height int) (*Screen, error) {
	sc := &Screen{
		lines: make([][]rune, height),
		dim: Dimension{
			Width:  width,
			Height: height,
		},
	}
	if err := sc.dim.Validate(); err != nil {
		return nil, err
	}
	for i := range sc.lines {
		sc.lines[i] = make([]rune, width)
	}
	sc.fillGrid()
	return sc, nil
}

func (s *Screen) Put(x, y int, cell Cell) error {
	var err error
	if !s.dim.Valid(x, y) {
		return fmt.Errorf("invalid coordinates (%d, %d)", x, y)

	}
	switch c := cell.(type) {
	case Content:
		err = s.putContent(x, y, c)
	case Segment:
		err = s.putConnector(x, y, c)
	default:
	}
	return err
}

func (s *Screen) Render(w io.Writer) error {
	ws := bufio.NewWriter(w)
	for i := range s.lines {
		for j := range s.lines[i] {
			_, err := ws.WriteRune(s.lines[i][j])
			if err != nil {
				return err
			}
		}
		if _, err := ws.WriteRune('\n'); err != nil {
			return err
		}
	}
	return ws.Flush()
}

func (s *Screen) putContent(x, y int, val Content) error {
	for i, r := range val.Value {
		s.writeChar(x+i, y, r)
	}
	return nil
}

func (s *Screen) putConnector(x, y int, seg Segment) error {
	if seg.Horizontal() {
		s.horizontalConnector(seg)
	} else {
		s.verticalConnector(seg)
	}
	return nil
}

func (s *Screen) verticalConnector(seg Segment) {
	if seg.DistanceY() == 1 {
		s.writeSymbol(seg.Start.X, seg.Start.Y, verticalBarAscii)
		return
	}
	var (
		start = seg.Start
		end   = seg.End
	)
	if start.BeforeY(end) {
		start, end = end, start
	}
	for y := start.Y; y >= end.Y; y-- {
		ch := verticalBarAscii
		if y == start.Y || y == end.Y {
			ch = connectBarAscii
		}
		s.writeSymbol(start.X, y, ch)
	}
}

func (s *Screen) horizontalConnector(seg Segment) {
	if seg.DistanceX() == 1 {
		s.writeSymbol(seg.Start.X, seg.Start.Y, horizontalBarAscii)
		return
	}
	var (
		start = seg.Start
		end   = seg.End
	)
	if end.BeforeX(start) {
		start, end = end, start
	}
	for x := start.X; x <= end.X; x++ {
		ch := horizontalBarAscii
		if x == start.X || x == end.X {
			ch = connectBarAscii
		}
		s.writeSymbol(x, start.Y, ch)
	}
}

func (s *Screen) writeChar(x, y int, char rune) {
	if !s.dim.Valid(x, y) {
		return
	}
	s.lines[y][x] = char
}

func (s *Screen) writeSymbol(x, y int, char rune) {
	// cell := c.cells[y*c.dim.Width+x]
	// if cell != nil {
	// 	b := cell.Rune()
	// 	if b == connectBarAscii && char == connectBarAscii {
	// 		return
	// 	}
	// 	if b == verticalBarAscii && char == horizontalBarAscii {
	// 		char = connectBarAscii
	// 	} else if b == horizontalBarAscii && char == verticalBarAscii {
	// 		char = connectBarAscii
	// 	}
	// }
	s.writeChar(x, y, char)
}

func (s *Screen) fillGrid() {
	for i := range s.lines {
		for j := range s.lines[i] {
			s.lines[i][j] = space
		}
	}
}
