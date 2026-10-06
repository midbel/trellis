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

func (o *SvgOptions) Metrics() Metrics {
	m := svgMetric{
		font:   svg.SansSerif,
		size:   12,
		margin: o.Margin,
	}
	return m
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

func (s *Svg) fill(x, y int, fill func(drawer) error) error {
	return fill(s)
}

func (s *Svg) draw(x, y int, cell Cell) error {
	var el svg.Element
	switch c := cell.(type) {
	case Content:
		el = s.createText(x, y, c)
	case Path:
		x, err := s.createPath(c)
		if err != nil {
			return err
		}
		el = x
	default:
		return fmt.Errorf("invalid cell type")
	}
	s.root.Add(el)
	return nil
}

func (s *Svg) createText(x, y int, c Content) svg.Element {
	t := svg.NewText(float64(x), float64(y), string(c.Value))
	var weight string
	if c.Style.IsBold() {
		weight = svg.WeightBold
	}
	t = t.Font(svg.NewFont(float64(c.Style.Size), c.Style.Family, weight))
	if c.Style.Color != "" {
		t = t.Fill(svg.Color(c.Style.Color))
	}
	return t
}

func (s *Svg) createPath(c Path) (svg.Element, error) {
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
	var paths []Path
	if s.opts.Orient == HorizontalLayout {
		paths = splitPathH(c)
	} else {
		paths = splitPathV(c)
	}
	if len(paths) <= 1 {
		return s.directPath(c)
	}
	var (
		fst = paths[0]
		lst = paths[len(paths)-1]
		p   = svg.NewPath()
	)

	p.MoveTo(float64(fst.Start.X), float64(fst.Start.Y))
	p.CurveTo(
		float64(lst.End.X),
		float64(lst.End.Y),
		float64(fst.End.X),
		float64(fst.End.Y),
		float64(lst.Start.X),
		float64(lst.Start.Y),
	)
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
