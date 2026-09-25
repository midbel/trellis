package trellis

import (
	"errors"
	"fmt"
)

var (
	ErrUnknown = errors.New("unknown")
	ErrOptions = errors.New("options should be provided")
)

const (
	PaddingS = 1
	PaddingM = 2
	PaddingL = 4
)

const (
	SpacingS = 1
	SpacingM = 2
	SpacingL = 4
	SpacingX = 8
)

const (
	DefaultSpacing = 2
	DefaultMargin  = 1
)

type Options interface {
	Layout() RenderOptions
	Format() Output
}

type ScreenOptions struct {
	RenderOptions
	Border          bool
	CoordinatesStep int
	Style           ConnectorStyle
}

func (o *ScreenOptions) Layout() RenderOptions {
	return o.RenderOptions
}

func (*ScreenOptions) Format() Output {
	return OutputScreen
}

type SvgOptions struct {
	RenderOptions
	Border          bool
	CoordinatesStep int
	Style           ConnectorStyle
	Path            PathStyle // manathan, direct, curve
}

func (o *SvgOptions) Layout() RenderOptions {
	return o.RenderOptions
}

func (*SvgOptions) Format() Output {
	return OutputSvg
}

type JsonOptions struct {
	RenderOptions
	Compact bool
}

func (o *JsonOptions) Layout() RenderOptions {
	return o.RenderOptions
}

func (*JsonOptions) Format() Output {
	return OutputJson
}

type XmlOptions struct {
	RenderOptions
	Compact bool
}

func (o *XmlOptions) Layout() RenderOptions {
	return o.RenderOptions
}

func (*XmlOptions) Format() Output {
	return OutputXml
}

type RenderOptions struct {
	Orient      Orientation
	Size        Dimension
	MinDepth    int
	MaxDepth    int
	Spacing     int // Distance between sibling allocation regions.
	Reverse     bool
	AlignY      Alignment
	AlignX      Alignment
	Margin      int // Space outside the node, mainly reserved for connectors
	Padding     int // Space inside the node's visual box
	PaddingChar string
	Transform   func(*Node, RenderOptions) Content
}

func (o *RenderOptions) Render(n *Node, opts RenderOptions) Content {
	if o.Transform == nil {
		return defaultRenderContent(n, opts)
	}
	return o.Transform(n, opts)
}

func (o *RenderOptions) Validate() error {
	if err := o.Size.Validate(); err != nil {
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
	if o.MinDepth < 0 {
		return fmt.Errorf("minimum depth can not be negative")
	}
	if o.MaxDepth < 0 {
		return fmt.Errorf("maximum depth can not be negative")
	}
	return nil
}

func (t *RenderOptions) ResetSize() {
	t.Size.Width = 0
	t.Size.Height = 0
}

func (t *RenderOptions) Clone() RenderOptions {
	x := *t
	return x
}

func (t *RenderOptions) Align() Alignment {
	if t.Orient == HorizontalLayout {
		return t.AlignX
	}
	return t.AlignY
}

func (t *RenderOptions) ApplyDefaults() {
	if t.Margin == 0 {
		t.Margin = SpacingL
	}
	if t.Spacing == 0 {
		t.Spacing = SpacingM
	}
	if t.Transform == nil {
		t.Transform = defaultRenderContent
	}
}

func defaultRenderContent(node *Node, opts RenderOptions) Content {
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

var defaultOptions = &RenderOptions{
	Spacing: DefaultSpacing,
	Margin:  DefaultMargin,
	Padding: PaddingS,
	AlignX:  AlignCenter,
	AlignY:  AlignCenter,
}

func unknown(what, value string) error {
	return fmt.Errorf("%s: %w %s", value, ErrUnknown, what)
}
