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

type Options interface {
	Layout() (RenderOptions, error)
	Format() Output
	applyDefaults()
}

type TableOptions struct {
	RenderOptions
}

func (o *TableOptions) Layout() (RenderOptions, error) {
	return o.RenderOptions, o.RenderOptions.Validate()
}

func (*TableOptions) Format() Output {
	return OutputTable
}

func (o *TableOptions) applyDefaults() {
	o.RenderOptions.applyDefaults()
}

type ScreenOptions struct {
	RenderOptions
	Border          bool
	CoordinatesStep int
	Style           ConnectorStyle
}

func (o *ScreenOptions) Layout() (RenderOptions, error) {
	return o.RenderOptions, o.RenderOptions.Validate()
}

func (*ScreenOptions) Format() Output {
	return OutputScreen
}

func (o *ScreenOptions) applyDefaults() {
	o.RenderOptions.applyDefaults()
	if o.Style == 0 {
		o.Style = ConnectorUnicode
	}
}

type SvgOptions struct {
	RenderOptions
	Border          bool
	CoordinatesStep int
	Style           ConnectorStyle
	Path            PathType // manathan, direct, curve
}

func (o *SvgOptions) Layout() (RenderOptions, error) {
	return o.RenderOptions, o.RenderOptions.Validate()
}

func (*SvgOptions) Format() Output {
	return OutputSvg
}

func (o *SvgOptions) applyDefaults() {
	o.RenderOptions.applyDefaults()
	if o.Style == 0 {
		o.Style = ConnectorUnicode
	}
	if o.Path == 0 {
		o.Path = ManathanPath
	}
}

type JsonOptions struct {
	RenderOptions
	Compact bool
}

func (o *JsonOptions) Layout() (RenderOptions, error) {
	return o.RenderOptions, o.RenderOptions.Validate()
}

func (*JsonOptions) Format() Output {
	return OutputJson
}

func (o *JsonOptions) applyDefaults() {
	o.RenderOptions.applyDefaults()
}

type XmlOptions struct {
	RenderOptions
	Compact bool
}

func (o *XmlOptions) Layout() (RenderOptions, error) {
	return o.RenderOptions, o.RenderOptions.Validate()
}

func (*XmlOptions) Format() Output {
	return OutputXml
}

func (o *XmlOptions) applyDefaults() {
	o.RenderOptions.applyDefaults()
}

type RenderOptions struct {
	AllocateMode Allocate
	Orient       Orientation
	Size         Dimension
	MinDepth     int
	MaxDepth     int
	Spacing      int // Distance between sibling allocation regions.
	Reverse      bool
	AlignY       Alignment
	AlignX       Alignment
	Margin       int // Space outside the node, mainly reserved for connectors
	Padding      int // Space inside the node's visual box
	PaddingChar  string
	Transform    func(*Node, RenderOptions) Content

	estimateMinSize int
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

func (t *RenderOptions) applyDefaults() {
	if t.Margin == 0 {
		t.Margin = SpacingL
	}
	if t.Spacing == 0 {
		t.Spacing = SpacingM
	}
	if t.Transform == nil {
		t.Transform = defaultRenderContent
	}
	if t.AllocateMode == 0 {
		t.AllocateMode = AllocateEqual
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

func unknown(what, value string) error {
	return fmt.Errorf("%s: %w %s", value, ErrUnknown, what)
}
