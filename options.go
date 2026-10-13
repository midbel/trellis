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
	Metrics(*Node) Metrics
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

type RenderOptions struct {
	AllocateMode Allocate
	Placement    Alignment
	Orient       Orientation
	Size         Dimension
	Path         PathType // manathan, direct, curve
	MinDepth     int
	MaxDepth     int
	Spacing      int // Distance between sibling allocation regions.
	Reverse      bool
	Transform    func(*Node, RenderOptions) Content

	estimateMinSize int
}

func (o *RenderOptions) Metrics(_ *Node) Metrics {
	return defaultMetric{}
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

func (t *RenderOptions) applyDefaults() {
	if t.Spacing == 0 {
		t.Spacing = SpacingM
	}
	if t.Transform == nil {
		t.Transform = defaultRenderContent
	}
	if t.AllocateMode == 0 {
		t.AllocateMode = AllocateEqual
	}
	if t.Orient == 0 {
		t.Orient = HorizontalLayout
	}
}

func defaultRenderContent(node *Node, opts RenderOptions) Content {
	return Content{
		Value: []rune(node.Value),
		Style: node.Style,
	}
}

func unknown(what, value string) error {
	return fmt.Errorf("%s: %w %s", value, ErrUnknown, what)
}
