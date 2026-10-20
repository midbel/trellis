package trellis

import (
	"io"
)

const (
	ansiEsc    = 0x1b
	ansiOpen   = '['
	ansiCmd    = 'm'
	ansiReset  = 0
	ansiBold   = '1'
	ansiItalic = '3'
	ansiSep    = ';'
)

func writeAnsiStyle(ws io.ByteWriter, color string, bold, italic bool) error {
	aw := newAnsiWriter(ws)
	if bold {
		aw.writeBold()
	}
	if italic {
		if bold {
			aw.writeSep()
		}
		aw.writeItalic()
	}
	if fst, snd := ansiColor(color); fst > 0 && snd > 0 {
		if bold || italic {
			aw.writeSep()
		}
		aw.writeByte(fst)
		aw.writeByte(snd)
	}
	aw.writeCmd()
	return aw.fail()
}

func writeAnsiClose(ws io.ByteWriter) error {
	aw := newAnsiWriter(ws)
	aw.writeCmd()
	return aw.fail()
}

func ansiColor(s string) (byte, byte) {
	switch s {
	case "black":
		return '3', '0'
	case "red":
		return '3', '1'
	case "green":
		return '3', '2'
	case "yellow":
		return '3', '3'
	case "blue":
		return '3', '4'
	case "magenta":
		return '3', '5'
	case "cyan":
		return '3', '6'
	case "white":
		return '3', '7'
	default:
		return 0, 0
	}
}

type ansiWriter struct {
	err    error
	writer io.ByteWriter
}

func newAnsiWriter(ws io.ByteWriter) *ansiWriter {
	aw := &ansiWriter{
		writer: ws,
	}
	aw.writeByte(ansiEsc)
	aw.writeByte(ansiOpen)
	return aw
}

func (w *ansiWriter) writeBold() {
	w.writeByte(ansiBold)
}

func (w *ansiWriter) writeItalic() {
	w.writeByte(ansiItalic)
}

func (w *ansiWriter) writeSep() {
	w.writeByte(ansiSep)
}

func (w *ansiWriter) writeReset() {
	w.writeByte(ansiReset)
}

func (w *ansiWriter) writeCmd() {
	w.writeByte(ansiCmd)
}

func (w *ansiWriter) fail() error {
	return w.err
}

func (w *ansiWriter) writeByte(b byte) {
	if w.err == nil {
		w.err = w.writer.WriteByte(b)
	}
}
