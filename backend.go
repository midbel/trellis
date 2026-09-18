package trellis

import (
	"bufio"
	"io"
)

const space = ' '

type Screen struct {
	bytes  [][]rune
	dim    Dimension
	filler rune
}

func NewScreen(width, height int) (*Screen, error) {
	sc := &Screen{
		bytes: make([][]rune, height),
		dim: Dimension{
			Width:  width,
			Height: height,
		},
		filler: space,
	}
	if err := sc.dim.Validate(); err != nil {
		return nil, err
	}
	for i := range sc.bytes {
		sc.bytes[i] = make([]rune, width)
	}
	return sc, nil
}

func (s *Screen) Put(x, y int, char rune) {
	if y >= 0 && y < len(s.bytes) {
		if x < 0 || x >= len(s.bytes[y]) {
			return
		}
		if char == 0 && s.bytes[y][x] == 0 {
			s.bytes[y][x] = s.filler
			return
		}
		s.bytes[y][x] = char
	}
}

func (s *Screen) Render(w io.Writer) error {
	ws := bufio.NewWriter(w)
	for i := range s.bytes {
		for j := range s.bytes[i] {
			_, err := ws.WriteRune(s.bytes[i][j])
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
