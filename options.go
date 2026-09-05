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
	DefaultSpacing = 2
	DefaultMargin  = 1
)

type LayoutOptions struct {
	Orient Orientation
	Dimension
	MinDepth int
	MaxDepth int
	Spacing  int // Distance between sibling allocation regions.
	Reverse  bool
	Render   func(*Node, *Options) Content
}

type RenderOptions struct {
	ShowCoordinates bool
	Border          bool
}

type StyleOptions struct {
	AlignY      Alignment
	AlignX      Alignment
	Margin      int // Space outside the node, mainly reserved for connectors
	Padding     int // Space inside the node's visual box
	PaddingChar string
	Style       ConnectorStyle
}

type Options struct {
	LayoutOptions
	RenderOptions
	StyleOptions
}

func (o *Options) Validate() error {
	if err := o.Dimension.Validate(); err != nil {
		return err
	}
	if o.Margin < 0 {
		return fmt.Errorf("margin can not be negative")
	}
	if o.Padding < 0 {
		return fmt.Errorf("padding can not be negative")
	}
	if o.Spacing < 0 {
		return fmt.Errorf("spacing can not be negative")
	}
	return nil
}

func prepareOptions(options *Options) (*Options, error) {
	opts := options
	if opts == nil {
		opts = defaultOptions.Clone()
	} else {
		opts = opts.Clone()
	}
	applyDefaults(opts)
	return opts, opts.Validate()
}

func applyDefaults(opts *Options) {
	if opts.Margin == 0 {
		opts.Margin++
	}
	if opts.Spacing == 0 {
		opts.Spacing++
	}
	if opts.Render == nil {
		opts.Render = defaultRenderContent
	}
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
		Spacing: DefaultSpacing,
	},
	StyleOptions: StyleOptions{
		Style:   ConnectorAscii,
		Margin:  DefaultMargin,
		Padding: PaddingS,
		AlignX:  AlignCenter,
		AlignY:  AlignCenter,
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

func (t *Options) Align() Alignment {
	if t.Orient == HorizontalLayout {
		return t.AlignX
	}
	return t.AlignY
}

func unknown(what, value string) error {
	return fmt.Errorf("%s: %w %s", value, ErrUnknown, what)
}
