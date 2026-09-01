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

func (s *Screen) Put(x, y int, content Content) {
	if y >= 0 && y < len(s.bytes) {
		if x < 0 || x >= len(s.bytes[y]) {
			return
		}
		if len(content.Value) == 0 && s.bytes[y][x] == 0 {
			s.bytes[y][x] = s.filler
			return
		}
		if content.kind == KindValue {
			s.putValue(x, y, content)
		} else if content.kind == KindConnector {
			s.putConnector(x, y, content)
		}
	}
}

func (s *Screen) putConnector(x, y int, content Content) {
	if x >= 0 && x < len(s.bytes[y]) {
		source := s.bytes[y][x]
		if source == connectBarAscii {
			return
		}
		content.Value[0] = replaceConnector(source, content.Value[0])
		s.bytes[y][x] = content.Value[0]
	}
}

func (s *Screen) putValue(x, y int, content Content) {
	for _, b := range content.Value {
		if x >= 0 && x < len(s.bytes[y]) {
			s.bytes[y][x] = b
			x++
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
