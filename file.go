package trellis

import (
	"fmt"
	"io"
	"strconv"

	"github.com/midbel/angle/xml"
	"github.com/midbel/curly"
)

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
		var v any
		if f.opts.Path == CurvePath {
			v = createCurveJsonPath(c, f.opts.Orient)
		} else if f.opts.Path == DirectPath {
			v = createDirectJsonPath(c, f.opts.Orient)
		} else {
			v = createManathanJsonPath(c, f.opts.Orient)
		}
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
		var v any
		if e.opts.Path == CurvePath {
			v = createCurveJsonPath(c, e.opts.Orient)
		} else if e.opts.Path == DirectPath {
			v = createDirectJsonPath(c, e.opts.Orient)
		} else {
			v = createManathanJsonPath(c, e.opts.Orient)
		}
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

func createCurveJsonPath(c Path, orient Orientation) any {
	var paths []Path
	if orient == HorizontalLayout {
		paths = splitPathH(c)
	} else {
		paths = splitPathV(c)
	}
	if len(paths) <= 1 {
		return createDirectJsonPath(c, orient)
	}
	g := map[string]any{
		"type":  "curve",
		"start": jsonPoint(c.Start),
		"end":   jsonPoint(c.End),
		"controls": []any{
			jsonPoint(paths[0].End),
			jsonPoint(paths[len(paths)-1].Start),
		},
	}
	return g
}

func createDirectJsonPath(c Path, orient Orientation) any {
	g := map[string]any{
		"type":  "direct",
		"start": jsonPoint(c.Start),
		"end":   jsonPoint(c.End),
	}
	return g
}

func createManathanJsonPath(c Path, orient Orientation) any {
	var paths []Path
	if orient == HorizontalLayout {
		paths = splitPathH(c)
	} else {
		paths = splitPathV(c)
	}
	var list []any
	for _, c := range paths {
		g := map[string]any{
			"start": jsonPoint(c.Start),
			"end":   jsonPoint(c.End),
		}
		list = append(list, g)
	}
	p := map[string]any{
		"type": "manathan",
		"runs": list,
	}
	return p
}

func jsonPoint(pt Point) any {
	p := map[string]any{
		"x": pt.X,
		"y": pt.Y,
	}
	return p
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
		if f.opts.Path == CurvePath {
			node = createCurveElement(c, f.opts.Orient)
		} else if f.opts.Path == DirectPath {
			node = createDirectElement(c, f.opts.Orient)
		} else {
			node = createManathanElement(c, f.opts.Orient)
		}
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
		if e.opts.Path == CurvePath {
			node = createCurveElement(c, e.opts.Orient)
		} else if e.opts.Path == DirectPath {
			node = createDirectElement(c, e.opts.Orient)
		} else {
			node = createManathanElement(c, e.opts.Orient)
		}
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

func createCurveElement(c Path, orient Orientation) xml.Node {
	var paths []Path
	if orient == HorizontalLayout {
		paths = splitPathH(c)
	} else {
		paths = splitPathV(c)
	}
	if len(paths) <= 1 {
		return createDirectElement(c, orient)
	}
	var (
		fst  = paths[0]
		lst  = paths[len(paths)-1]
		ctrl = xml.NewElement(xml.NewName("controls"))
	)
	ctrl.Children = []xml.Node{
		xmlPoint("ctrl1", fst.End),
		xmlPoint("ctrl2", lst.Start),
	}

	el := xml.NewElement(xml.NewName("path"))
	attr := xml.NewAttribute(xml.NewName("type"), "curve")
	el.Attributes = append(el.Attributes, attr)
	el.Children = []xml.Node{
		xmlPoint("start", c.Start),
		xmlPoint("end", c.End),
		ctrl,
	}
	return el
}

func createDirectElement(c Path, orient Orientation) xml.Node {
	el := xml.NewElement(xml.NewName("path"))
	attr := xml.NewAttribute(xml.NewName("type"), "direct")
	el.Attributes = append(el.Attributes, attr)
	el.Children = []xml.Node{
		xmlPoint("start", c.Start),
		xmlPoint("end", c.End),
	}
	return el
}

func createManathanElement(c Path, orient Orientation) xml.Node {
	el := xml.NewElement(xml.NewName("path"))
	attr := xml.NewAttribute(xml.NewName("type"), "manathan")
	el.Attributes = append(el.Attributes, attr)

	var paths []Path
	if orient == HorizontalLayout {
		paths = splitPathH(c)
	} else {
		paths = splitPathV(c)
	}
	for i := range paths {
		n := createManathanRun(paths[i])
		el.Children = append(el.Children, n)
	}
	return el
}

func createManathanRun(c Path) xml.Node {
	run := xml.NewElement(xml.NewName("run"))
	run.Children = []xml.Node{
		xmlPoint("start", c.Start),
		xmlPoint("end", c.End),
	}
	return run
}

func xmlPoint(name string, pt Point) xml.Node {
	el := xml.NewElement(xml.NewName(name))
	el.Attributes = []xml.Attribute{
		xml.NewAttribute(xml.NewName("x"), strconv.Itoa(pt.X)),
		xml.NewAttribute(xml.NewName("y"), strconv.Itoa(pt.Y)),
	}
	return el
}
