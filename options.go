package trellis

import (
	"errors"
	"fmt"

	"github.com/midbel/angle/svg"
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
	Format() Output
	applyDefaults()
}

type OutputOptions struct {
	Mode            Output
	Compact         bool
	CoordinatesStep int
	Border          bool
	Connector       ConnectorStyle
	Path            PathType
}

func (o *OutputOptions) Metrics(n *Node) Metrics {
	if o.Mode == OutputSvg {
		m := svgMetric{
			font: n.Style.Family,
			size: n.Style.Size,
		}
		if m.size == 0 {
			m.size = svg.DefaultFontSize
		}
		return m
	}
	return defaultMetric{}
}

func (o OutputOptions) toSvg() (SvgOptions, error) {
	opts := SvgOptions{
		Border:          o.Border,
		Style:           o.Connector,
		CoordinatesStep: o.CoordinatesStep,
		Path:            o.Path,
	}
	if opts.CoordinatesStep < 0 {
		return opts, fmt.Errorf("negative value not allow for ticks")
	}
	opts.applyDefaults()
	return opts, nil
}

func (o OutputOptions) toScreen() (ScreenOptions, error) {
	opts := ScreenOptions{
		Border:          o.Border,
		Style:           o.Connector,
		CoordinatesStep: o.CoordinatesStep,
		Path:            o.Path,
	}
	if opts.CoordinatesStep < 0 {
		return opts, fmt.Errorf("negative value not allow for ticks")
	}
	opts.applyDefaults()
	return opts, nil
}

func (o OutputOptions) toXml() (XmlOptions, error) {
	opts := XmlOptions{
		Compact: o.Compact,
		Path:    o.Path,
	}
	return opts, nil
}

func (o OutputOptions) toJson() (JsonOptions, error) {
	opts := JsonOptions{
		Compact: o.Compact,
		Path:    o.Path,
	}
	return opts, nil
}

type RenderOptions struct {
	AllocateMode Allocate
	Placement    Alignment
	Orient       Orientation
	Size         Dimension
	MinDepth     int
	MaxDepth     int
	Spacing      int // Distance between sibling allocation regions.
	Reverse      bool

	estimateMinSize int
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
	if t.AllocateMode == 0 {
		t.AllocateMode = AllocateEqual
	}
	if t.Orient == 0 {
		t.Orient = HorizontalLayout
	}
}

func unknown(what, value string) error {
	return fmt.Errorf("%s: %w %s", value, ErrUnknown, what)
}
