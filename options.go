package trellis

import (
	"errors"
	"fmt"
)

var ErrUnknown = errors.New("unknown")

const (
	PaddingS = 1
	PaddingM = 2
	PaddingL = 4
)

const (
	DefaultSiblingGap = 2
	DefaultLevelGap   = 5
)

type LayoutOptions struct {
	Orient     Orientation
	Width      int
	Height     int
	MinDepth   int
	MaxDepth   int
	LevelGap   int
	SiblingGap int
	Anchor     Alignment
	Reverse    bool
	Render     func(*Node, *Options) Content
}

type RenderOptions struct {
	ShowCoordinates bool
	Border          bool
}

type StyleOptions struct {
	Align       Alignment
	AlignY      Alignment
	AlignX      Alignment
	Padding     int
	PaddingChar string
	Style       ConnectorStyle
}

type Options struct {
	LayoutOptions
	RenderOptions
	StyleOptions
}

func prepareOptions(options *Options) *Options {
	opts := options
	if opts == nil {
		opts = defaultOptions.Clone()
	} else {
		opts = opts.Clone()
	}
	if opts.LevelGap == 0 {
		opts.LevelGap++
	}
	if opts.SiblingGap == 0 {
		opts.SiblingGap++
	}
	if opts.Render == nil {
		opts.Render = defaultRenderContent
	}
	return opts
}

func defaultRenderContent(node *Node, opts *Options) Content {
	value := []byte(node.Value)
	if opts.Padding > 0 {
		var (
			pad = make([]byte, opts.Padding)
			tmp = make([]byte, 0, len(value))
		)
		for i := range pad {
			pad[i] = ' '
		}
		tmp = append(tmp, pad...)
		tmp = append(tmp, value...)
		tmp = append(tmp, pad...)

		value = tmp
	}
	return Content{
		Value: value,
	}
}

var defaultOptions = &Options{
	LayoutOptions: LayoutOptions{
		SiblingGap: DefaultSiblingGap,
		LevelGap:   DefaultLevelGap,
		Anchor:     AlignCenter,
	},
	StyleOptions: StyleOptions{
		Style:   ConnectorAscii,
		Padding: PaddingS,
		Align:   AlignCenter,
	},
	RenderOptions: RenderOptions{
		Border: true,
	},
}

func (t *Options) Clone() *Options {
	x := *t
	return &x
}

func (t *Options) borderWidth() int {
	if t.Border {
		return 1
	}
	return 0
}

func (t *Options) paddedValue(str string) []byte {
	if t.Padding <= 0 {
		return []byte(str)
	}
	var (
		char  = byte(' ')
		value = []byte(str)
		pad   = make([]byte, t.Padding)
		tmp   = make([]byte, 0, t.Padding+len(str))
	)
	if t.PaddingChar != "" && len(t.PaddingChar) == 1 {
		char = t.PaddingChar[0]
	}
	for i := range pad {
		pad[i] = char
	}
	tmp = append(tmp, pad...)
	tmp = append(tmp, value...)
	tmp = append(tmp, pad...)
	return tmp
}

func (t *Options) levelGaps() int {
	return t.LevelGap + t.LevelGap
}

func (t *Options) siblingsGaps() int {
	return t.SiblingGap + t.SiblingGap
}

func unknown(what, value string) error {
	return fmt.Errorf("%s: %w %s", value, ErrUnknown, what)
}
