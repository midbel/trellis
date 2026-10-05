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
	ws.WriteByte(ansiEsc)
	ws.WriteByte(ansiOpen)
	if bold {
		ws.WriteByte(ansiBold)
	}
	if italic {
		if bold {
			ws.WriteByte(ansiSep)
		}
		ws.WriteByte(ansiItalic)
	}
	if fst, snd := ansiColor(color); fst > 0 && snd > 0 {
		if bold || italic {
			ws.WriteByte(ansiSep)
		}
		ws.WriteByte(fst)
		ws.WriteByte(snd)
	}
	ws.WriteByte(ansiCmd)
	return nil
}

func writeAnsiColor(ws io.ByteWriter, color string) error {
	fst, snd := ansiColor(color)
	if fst == 0 && snd == 0 {
		return nil
	}
	ws.WriteByte(ansiEsc)
	ws.WriteByte(ansiOpen)
	ws.WriteByte(fst)
	ws.WriteByte(snd)
	ws.WriteByte(ansiCmd)
	return nil
}

func writeAnsiClose(ws io.ByteWriter) error {
	ws.WriteByte(ansiEsc)
	ws.WriteByte(ansiOpen)
	ws.WriteByte(ansiReset)
	ws.WriteByte(ansiCmd)
	return nil
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
