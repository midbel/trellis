package trellis

import (
	"errors"
	"fmt"
)

var ErrUnknown = errors.New("unknown")

const (
	PaddingS = 1 << iota
	PaddingM
	PaddingL
)

const (
	SpacingS = 1 << iota
	SpacingM
	SpacingL
	SpacingX
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
	CoordinatesStep int
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
	Output Output
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
	if o.CoordinatesStep < 0 {
		return fmt.Errorf("coordinates step can not be negative")
	}
	if o.MinDepth < 0 {
		return fmt.Errorf("minimum depth can not be negative")
	}
	if o.MaxDepth < 0 {
		return fmt.Errorf("maximum depth can not be negative")
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
	value := []rune(node.Value)
	if opts.Padding > 0 {
		var (
			pad  = make([]rune, opts.Padding)
			tmp  = make([]rune, 0, len(value))
			char = rune(' ')
		)
		if len(opts.PaddingChar) == 1 {
			raw := []rune(opts.PaddingChar)
			char = raw[0]
		}
		for i := range pad {
			pad[i] = char
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

func (t *Options) Align() Alignment {
	if t.Orient == HorizontalLayout {
		return t.AlignX
	}
	return t.AlignY
}

func unknown(what, value string) error {
	return fmt.Errorf("%s: %w %s", value, ErrUnknown, what)
}
