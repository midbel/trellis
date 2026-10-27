package codec

import (
	"github.com/midbel/trellis"
)

type OptionsBag struct {
	Type         trellis.Output
	Orient       trellis.Orientation
	AllocateMode trellis.Allocate
	Placement    trellis.Alignment
	Width        int
	Height       int
	MinDepth     int
	MaxDepth     int
	Spacing      int // Distance between sibling allocation regions.
	Reverse      bool

	Border          bool
	CoordinatesStep int
	Style           trellis.ConnectorStyle
	Path            trellis.PathType // manathan, direct, curve

	Compact bool

	Global trellis.Style
}

func defaultBag() *OptionsBag {
	return new(OptionsBag)
}

func (b *OptionsBag) RenderOptions() *trellis.RenderOptions {
	return &trellis.RenderOptions{
		Output:       b.Type,
		AllocateMode: b.AllocateMode,
		Placement:    b.Placement,
		Orient:       b.Orient,
		Size:         trellis.NewDimension(b.Width, b.Height),
		MinDepth:     b.MinDepth,
		MaxDepth:     b.MaxDepth,
		Spacing:      b.Spacing,
		Reverse:      b.Reverse,
	}
}

func (b *OptionsBag) OutputOptions() *trellis.OutputOptions {
	return &trellis.OutputOptions{
		Mode:            b.Type,
		Compact:         b.Compact,
		CoordinatesStep: b.CoordinatesStep,
		Border:          b.Border,
		Connector:       b.Style,
		Path:            b.Path,
	}
}
