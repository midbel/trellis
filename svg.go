package trellis

import (
	"fmt"
	"io"

	"github.com/midbel/angle/svg"
)

type SvgOptions struct {
	RenderOptions
	Border          bool
	CoordinatesStep int
	Style           ConnectorStyle
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

type Svg struct {
	root *svg.Document
	opts SvgOptions
}

func NewSvg(opts SvgOptions) (View, error) {
	doc := svg.NewDocument(float64(opts.Size.Width), float64(opts.Size.Height))
	return &Svg{
		root: doc,
		opts: opts,
	}, nil
}

func (s *Svg) Render(w io.Writer) error {
	return s.root.Render(w)
}

func (s *Svg) draw(x, y int, cell Cell) error {
	var el svg.Element
	switch c := cell.(type) {
	case Content:
		el = svg.NewText(float64(x), float64(y), string(c.Value))
	case Path:
		x, err := s.putPath(c)
		if err != nil {
			return err
		}
		el = x
	default:
		return fmt.Errorf("invalid cell type")
	}
	s.root.Append(el)
	return nil
}

func (s *Svg) fill(x, y int, fill func(drawer) error) error {
	return fill(s)
}

func (s *Svg) putPath(c Path) (svg.Element, error) {
	switch s.opts.Path {
	case ManathanPath:
		return s.manathanPath(c)
	case DirectPath:
		return s.directPath(c)
	case CurvePath:
		return s.curvePath(c)
	default:
		return nil, fmt.Errorf("unsupported path type")
	}
}

func (s *Svg) curvePath(c Path) (svg.Element, error) {
	p := svg.NewPath()
	// p.MoveTo(float64(c.X()), float64(c.Y()))
	// if len(c.Paths) == 1 {
	// 	p.LineTo(float64(c.Paths[0].End.X), float64(c.Paths[0].End.Y))
	// 	return p, nil
	// }
	// if n := len(c.Paths) - 1; n > 0 {
	// 	x := c.Paths[n].End.X
	// 	y := c.Paths[n].End.Y
	// 	d := x - c.X()
	// 	p.CurveTo(
	// 		float64(x),
	// 		float64(y),
	// 		float64(c.X()+d),
	// 		float64(c.Y()),
	// 		float64(x-d),
	// 		float64(y),
	// 	)
	// }
	return p, nil
}

func (s *Svg) directPath(c Path) (svg.Element, error) {
	p := svg.NewLine(
		float64(c.Start.X),
		float64(c.Start.Y),
		float64(c.End.X),
		float64(c.End.Y),
	)
	return p, nil
}

func (s *Svg) manathanPath(c Path) (svg.Element, error) {
	var paths []Path
	if s.opts.Orient == HorizontalLayout {
		paths = splitPathH(c)
	} else {
		paths = splitPathV(c)
	}
	p := svg.NewPath()
	for i, s := range paths {
		if i == 0 {
			p.MoveTo(float64(s.Start.X), float64(s.Start.Y))
		} else {
			p.LineTo(float64(s.Start.X), float64(s.Start.Y))
		}
		p.LineTo(float64(s.End.X), float64(s.End.Y))
	}
	return p, nil
}
