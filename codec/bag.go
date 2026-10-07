package codec

import (
	"fmt"

	"github.com/midbel/trellis"
)

type OptionsBag struct {
	Type         trellis.Output
	Orient       trellis.Orientation
	AllocateMode trellis.Allocate
	Width        int
	Height       int
	MinDepth     int
	MaxDepth     int
	Spacing      int // Distance between sibling allocation regions.
	Reverse      bool
	AlignY       trellis.Alignment
	AlignX       trellis.Alignment
	Margin       int // Space outside the node, mainly reserved for connectors
	Padding      int // Space inside the node's visual box

	Border          bool
	CoordinatesStep int
	Style           trellis.ConnectorStyle
	Path            trellis.PathType // manathan, direct, curve

	Compact bool
}

func defaultBag() *OptionsBag {
	return new(OptionsBag)
}

func (b *OptionsBag) Build() (trellis.Options, error) {
	base := trellis.RenderOptions{
		AllocateMode: b.AllocateMode,
		Orient:       b.Orient,
		Size:         trellis.NewDimension(b.Width, b.Height),
		MinDepth:     b.MinDepth,
		MaxDepth:     b.MaxDepth,
		Spacing:      b.Spacing,
		Reverse:      b.Reverse,
		AlignY:       b.AlignY,
		AlignX:       b.AlignX,
		Margin:       b.Margin,
		Padding:      b.Padding,
		Path:         b.Path,
	}
	var opts trellis.Options
	switch b.Type {
	case trellis.OutputTable:
		opts = &trellis.TableOptions{
			RenderOptions: base,
		}
	case trellis.OutputSvg:
		opts = &trellis.SvgOptions{
			RenderOptions:   base,
			Style:           b.Style,
			CoordinatesStep: b.CoordinatesStep,
		}
	case trellis.OutputScreen:
		opts = &trellis.ScreenOptions{
			RenderOptions:   base,
			Border:          b.Border,
			Style:           b.Style,
			CoordinatesStep: b.CoordinatesStep,
		}
	case trellis.OutputXml:
		opts = &trellis.XmlOptions{
			RenderOptions: base,
			Compact:       b.Compact,
		}
	case trellis.OutputJson:
		opts = &trellis.JsonOptions{
			RenderOptions: base,
			Compact:       b.Compact,
		}
	default:
		return nil, fmt.Errorf("unsupported output type")
	}
	return opts, nil
}
