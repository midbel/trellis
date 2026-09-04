package trellis

import (
	"io"
)

const space = ' '

type Screen struct {
	bytes  [][]byte
	dim    Dimension
	filler byte
}

func NewScreen(width, height int) *Screen {
	sc := &Screen{
		bytes: make([][]byte, height),
		dim: Dimension{
			Width:  width,
			Height: height,
		},
		filler: space,
	}
	for i := range sc.bytes {
		sc.bytes[i] = make([]byte, width)
	}
	return sc
}

func (s *Screen) Put(x, y int, b byte) {
	if y >= 0 && y < len(s.bytes) {
		if x < 0 || x >= len(s.bytes[y]) {
			return
		}
		if b == 0 && s.bytes[y][x] == 0 {
			s.bytes[y][x] = s.filler
			return
		}
		s.bytes[y][x] = b
	}
}

func (s *Screen) Render(w io.Writer) error {
	for i := range s.bytes {
		_, err := w.Write(s.bytes[i])
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte{'\n'}); err != nil {
			return err
		}
	}
	return nil
}
