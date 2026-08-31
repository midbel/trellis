package trellis

import (
	"io"
)

type Screen struct {
	bytes [][]byte
}

func NewScreen(width, height int) *Screen {
	sc := &Screen{
		bytes: make([][]byte, height),
	}
	for i := range sc.bytes {
		sc.bytes[i] = make([]byte, width)
	}
	return sc
}

func (s *Screen) Put(x, y int, content Content) {
	if y >= 0 && y < len(s.bytes) {
		if x < 0 || x >= len(s.bytes[y]) {
			return
		}
		if len(content.Value) == 0 && s.bytes[y][x] == 0 {
			s.bytes[y][x] = ' '
			return
		}
		for _, b := range content.Value {
			if x >= 0 && x < len(s.bytes[y]) {
				s.bytes[y][x] = b
				x++
			}
		}
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
