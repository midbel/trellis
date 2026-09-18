package trellis

import (
	"bufio"
	"io"
)

const space = ' '

type Screen struct {
	lines  [][]rune
	dim    Dimension
	filler rune
}

func NewScreen(width, height int) (*Screen, error) {
	sc := &Screen{
		lines: make([][]rune, height),
		dim: Dimension{
			Width:  width,
			Height: height,
		},
		filler: space,
	}
	if err := sc.dim.Validate(); err != nil {
		return nil, err
	}
	for i := range sc.lines {
		sc.lines[i] = make([]rune, width)
	}
	return sc, nil
}

func (s *Screen) Put(x, y int, char rune) error {
	if y >= 0 && y < len(s.lines) {
		if x < 0 || x >= len(s.lines[y]) {
			return nil
		}
		if char == 0 && s.lines[y][x] == 0 {
			s.lines[y][x] = s.filler
			return nil
		}
		s.lines[y][x] = char
	}
	return nil
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
