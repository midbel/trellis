package trellis

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/midbel/angle/svg"
	"github.com/midbel/angle/xml"
	"github.com/midbel/curly"
)

const space = ' '

type drawer interface {
	draw(int, int, Cell) error
}

type View interface {
	Render(io.Writer) error
	fill(int, int, func(drawer) error) error
	drawer
}

type JsonFile struct {
	root map[string]any
	opts JsonOptions
}

func NewJson(opts JsonOptions) (View, error) {
	j := &JsonFile{
		root: make(map[string]any),
		opts: opts,
	}
	j.root["width"] = opts.Size.Width
	j.root["height"] = opts.Size.Height
	j.root["orientation"] = opts.Orient.String()
	j.root["cells"] = []any{}
	j.root["connectors"] = []any{}
	j.root["canvas"] = []any{}
	return j, nil
}

func (f *JsonFile) Render(w io.Writer) error {
	var ws *curly.Writer
	if f.opts.Compact {
		ws = curly.Compact(w)
	} else {
		ws = curly.NewWriter(w)
	}
	return ws.Write(f.root)
}

func (f *JsonFile) draw(x, y int, cell Cell) error {
	switch c := cell.(type) {
	case Content:
		v := createJsonContent(x, y, c)
		vs, ok := f.root["cells"].([]any)
		if ok {
			f.root["cells"] = append(vs, v)
		}
	case Path:
		v := createJsonPath(x, y, c)
		vs, ok := f.root["paths"].([]any)
		if ok {
			f.root["paths"] = append(vs, v)
		}
	default:
		return fmt.Errorf("invalid cell type")
	}
	return nil
}

func (f *JsonFile) fill(x, y int, fill func(drawer) error) error {
	j := newJsonElement(x, y, f.opts)
	if err := fill(j); err != nil {
		return err
	}
	vs, ok := f.root["canvas"].([]any)
	if ok {
		f.root["canvas"] = append(vs, j.canvas)
	}
	return nil
}

type jsonElement struct {
	canvas map[string]any
	opts   JsonOptions
}

func newJsonElement(x, y int, opts JsonOptions) *jsonElement {
	e := &jsonElement{
		canvas: make(map[string]any),
		opts:   opts,
	}
	e.canvas["x"] = x
	e.canvas["y"] = y
	e.canvas["cells"] = []any{}
	e.canvas["paths"] = []any{}
	e.canvas["canvas"] = []any{}
	return e
}

func (e *jsonElement) draw(x, y int, cell Cell) error {
	switch c := cell.(type) {
	case Content:
		v := createJsonContent(x, y, c)
		vs, ok := e.canvas["cells"].([]any)
		if ok {
			e.canvas["cells"] = append(vs, v)
		}
	case Path:
		v := createJsonPath(x, y, c)
		vs, ok := e.canvas["paths"].([]any)
		if ok {
			e.canvas["paths"] = append(vs, v)
		}
	default:
		return fmt.Errorf("invalid cell type")
	}
	return nil
}

func createJsonContent(x, y int, c Content) any {
	v := map[string]any{
		"content": string(c.Value),
		"x":       x,
		"y":       y,
	}
	return v
}

func createJsonPath(x, y int, c Path) any {
	s := map[string]any{
		"x": c.Start.X,
		"y": c.Start.Y,
	}
	e := map[string]any{
		"x": c.End.X,
		"y": c.End.Y,
	}
	g := map[string]any{
		"start": s,
		"end":   e,
	}
	return g
}

type XmlFile struct {
	root *xml.Element
	opts XmlOptions
}

func NewXml(opts XmlOptions) (View, error) {
	el := xml.Element{
		Name: xml.NewName("canvas"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("width"), strconv.Itoa(opts.Size.Width)),
			xml.NewAttribute(xml.NewName("height"), strconv.Itoa(opts.Size.Height)),
			xml.NewAttribute(xml.NewName("orientation"), opts.Orient.String()),
		},
	}
	return &XmlFile{
		root: &el,
		opts: opts,
	}, nil
}

func (f *XmlFile) Render(w io.Writer) error {
	var (
		doc = xml.NewDocument(f.root)
		enc = xml.NewEncoder(w)
	)
	enc.SetCompact(f.opts.Compact)
	return enc.Encode(doc)
}

func (f *XmlFile) draw(x, y int, cell Cell) error {
	var node xml.Node
	switch c := cell.(type) {
	case Content:
		node = createElementForContent(x, y, c)
	case Path:
		node = createElementForPath(x, y, c)
	default:
		return fmt.Errorf("invalid cell type")
	}
	if node != nil {
		f.root.Children = append(f.root.Children, node)
	}
	return nil
}

func (f *XmlFile) fill(x, y int, fill func(drawer) error) error {
	el := newXmlElement(x, y, f.opts)
	if err := fill(el); err != nil {
		return err
	}
	f.root.Children = append(f.root.Children, el.root)
	return nil
}

type xmlElement struct {
	root *xml.Element
	opts XmlOptions
}

func newXmlElement(x, y int, opts XmlOptions) *xmlElement {
	el := &xml.Element{
		Name: xml.NewName("canvas"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("x"), strconv.Itoa(x)),
			xml.NewAttribute(xml.NewName("y"), strconv.Itoa(y)),
		},
	}
	return &xmlElement{
		root: el,
		opts: opts,
	}
}

func (e *xmlElement) draw(x, y int, cell Cell) error {
	var node xml.Node
	switch c := cell.(type) {
	case Content:
		node = createElementForContent(x, y, c)
	case Path:
		node = createElementForPath(x, y, c)
	default:
		return fmt.Errorf("invalid cell type")
	}
	if node != nil {
		e.root.Children = append(e.root.Children, node)
	}
	return nil
}

func createElementForContent(x, y int, val Content) xml.Node {
	el := xml.Element{
		Name: xml.NewName("content"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("x"), strconv.Itoa(x)),
			xml.NewAttribute(xml.NewName("y"), strconv.Itoa(y)),
		},
		Children: []xml.Node{
			xml.NewText(string(val.Value)),
		},
	}
	return &el
}

func createElementForPath(x, y int, c Path) xml.Node {
	start := &xml.Element{
		Name: xml.NewName("start"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("x"), strconv.Itoa(c.Start.X)),
			xml.NewAttribute(xml.NewName("y"), strconv.Itoa(c.Start.Y)),
		},
	}
	end := &xml.Element{
		Name: xml.NewName("end"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("x"), strconv.Itoa(c.End.X)),
			xml.NewAttribute(xml.NewName("y"), strconv.Itoa(c.End.Y)),
		},
	}

	el := &xml.Element{
		Name: xml.NewName("path"),
		Children: []xml.Node{
			start,
			end,
		},
	}
	return el
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
	p := svg.NewPath()
	// p.MoveTo(float64(c.X()), float64(c.Y()))
	// if len(c.Paths) == 1 {
	// 	p.LineTo(float64(c.Paths[0].End.X), float64(c.Paths[0].End.Y))
	// 	return p, nil
	// }
	// if n := len(c.Paths) - 1; n > 0 {
	// 	p.LineTo(float64(c.Paths[n].End.X), float64(c.Paths[n].End.Y))
	// }
	return p, nil
}

func (s *Svg) manathanPath(c Path) (svg.Element, error) {
	p := svg.NewPath()
	// for i, s := range c.Paths {
	// 	if i == 0 {
	// 		p.MoveTo(float64(s.Start.X), float64(s.Start.Y))
	// 	} else {
	// 		p.LineTo(float64(s.Start.X), float64(s.Start.Y))
	// 	}
	// 	p.LineTo(float64(s.End.X), float64(s.End.Y))
	// }
	return p, nil
}

type Table struct {
	cells []Placement
}

func NewTable() (View, error) {
	t := &Table{}
	return t, nil
}

func (t *Table) Render(w io.Writer) error {
	fmt.Fprintf(w, "%-16s | %6s | %6s", "Value", "X", "Y")
	fmt.Fprintln(w)
	for _, p := range t.cells {
		var lines []any
		switch c := p.Cell.(type) {
		case Content:
			lines = t.getContentInfo(p.X, p.Y, c)
		case Path:
			lines = t.getPathInfo(p.X, p.Y, c)
		default:
			return fmt.Errorf("invalid cell type")
		}
		if len(lines) == 0 {
			continue
		}
		fmt.Fprintf(w, "%-16s | %6d | %6d", lines...)
		fmt.Fprintln(w)
	}
	return nil
}

func (t *Table) draw(x, y int, cell Cell) error {
	p := Placement{
		Point: NewPoint(x, y),
		Cell:  cell,
	}
	t.cells = append(t.cells, p)
	return nil
}

func (t *Table) fill(x, y int, fill func(drawer) error) error {
	return fill(t)
}

func (t *Table) getContentInfo(x, y int, c Content) []any {
	return []any{
		strings.TrimSpace(string(c.Value)),
		x,
		y,
	}
}

func (t *Table) getPathInfo(x, y int, c Path) []any {
	return []any{
		"path",
		c.X(),
		c.Y(),
	}
}
