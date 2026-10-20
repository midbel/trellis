package codec

import (
	"errors"
	"fmt"
	"io"

	"github.com/midbel/angle/xml"
	"github.com/midbel/curly"
	"github.com/midbel/sexpr"
	"github.com/midbel/trellis"
)

var (
	ErrFormat    = errors.New("format not supported")
	ErrIdent     = errors.New("identifier expected")
	ErrDirective = errors.New("unsupported directive")
	ErrType      = errors.New("invalid value type")
	ErrRoot      = errors.New("no root in document")
)

type Format uint8

const (
	FormatSexpr Format = iota
	FormatDot
	FormatJson
	FormatXml
)

type Options struct {
	Format Format
}

var dispatch = map[Format]func(io.Reader) (*TreeSpec, error){
	FormatSexpr: treeFromSexpr,
	FormatDot:   treeFromDot,
	FormatJson:  treeFromJson,
	FormatXml:   treeFromXml,
}

type TreeSpec struct {
	*trellis.Node
	Options *OptionsBag
}

func (t *TreeSpec) Build() (trellis.Options, error) {
	return t.Options.Build()
}

func Tree(r io.Reader, opts Options) (*TreeSpec, error) {
	fn, ok := dispatch[opts.Format]
	if !ok {
		return nil, ErrFormat
	}
	return fn(r)
}

func treeFromDot(r io.Reader) (*TreeSpec, error) {
	return nil, ErrFormat
}

func treeFromJson(r io.Reader) (*TreeSpec, error) {
	data, err := curly.Decode5(r)
	if err != nil {
		return nil, err
	}
	return buildSpecFromJson(data)
}

func buildSpecFromJson(data any) (*TreeSpec, error) {
	root, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected object")
	}
	var (
		source  any
		options = defaultBag()
	)
	if err := getOptionsFromJson(root["rendering"], options); err != nil {
		return nil, err
	}
	if d, ok := root["data"]; ok {
		source = d
	} else {
		source = data
	}
	tree, err := buildNodeFromJson(source, options.Global)
	if err != nil {
		return nil, err
	}
	spec := &TreeSpec{
		Node:    tree,
		Options: options,
	}
	return spec, nil
}

func buildNodeFromJson(data any, global trellis.Style) (*trellis.Node, error) {
	root, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected object")
	}
	name, ok := root["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing name property")
	}
	n := &trellis.Node{
		Value: name,
		Style: global,
	}
	if err := getStyleFromJson(root["style"], &n.Style); err != nil {
		return nil, err
	}
	if children, ok := root["children"].([]any); ok {
		for _, c := range children {
			sub, err := buildNodeFromJson(c, global)
			if err != nil {
				return nil, err
			}
			if sub != nil {
				n.Nodes = append(n.Nodes, sub)
			}
		}
	} else {
		return nil, fmt.Errorf("children to be an array")
	}
	return n, nil
}

func getOptionsFromJson(data any, options *OptionsBag) error {
	if data == nil {
		return nil
	}
	opts, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("options should be an object")
	}
	if style, ok := opts["style"]; ok {
		if err := getStyleFromJson(style, &options.Global); err != nil {
			return err
		}
		delete(opts, "style")
	}
	setters := makeSetters(options)
	for k, o := range opts {
		fn, ok := setters[k]
		if !ok {
			return fmt.Errorf("%s: unknown option", k)
		}
		if err := fn(o); err != nil {
			return fmt.Errorf("%s: option(json): %w - %s", k, err, o)
		}
	}
	return nil
}

func getStyleFromJson(data any, style *trellis.Style) error {
	if data == nil {
		return nil
	}
	y, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("style should be given as object")
	}
	setters := makeStyleSetters(style)
	for k, v := range y {
		fn, ok := setters[k]
		if !ok {
			return fmt.Errorf("%s: unknown style property", k)
		}
		if err := fn(v); err != nil {
			return fmt.Errorf("%s: style(json): %w - %v", k, err, v)
		}
	}
	return nil
}

func treeFromXml(r io.Reader) (*TreeSpec, error) {
	doc, err := xml.Decode(r)
	if err != nil {
		return nil, err
	}
	root, err := doc.Root()
	if err != nil {
		return nil, err
	}
	options := defaultBag()
	if err := getOptionsFromXml(root, options); err != nil {
		return nil, err
	}
	top, err := getNodeFromRoot(root)
	if err != nil {
		return nil, err
	}
	tree, err := buildNodeFromXml(top, options.Global)
	if err != nil {
		return nil, err
	}
	spec := &TreeSpec{
		Node: tree,
		Options: options,
	}
	return spec, nil
}

func buildNodeFromXml(node xml.Node, global trellis.Style) (*trellis.Node, error) {
	el, ok := node.(*xml.Element)
	if !ok {
		return nil, fmt.Errorf("element expected")
	}

	n := &trellis.Node{
		Style: global,
	}
	if err := setNodeValue(n, el); err != nil {
		return nil, err
	}
	for _, c := range el.ElementsByName("node") {
		child, err := buildNodeFromXml(c, global)
		if err != nil {
			return nil, err
		}
		n.Nodes = append(n.Nodes, child)
	}
	return n, nil
}

func getNodeFromRoot(node xml.Node) (xml.Node, error) {
	el, ok := node.(*xml.Element)
	if !ok {
		return nil, fmt.Errorf("element expected")
	}
	return el.ElementByName("node")
}

func setNodeValue(node *trellis.Node, el xml.Node) error {
	name, ok := el.(*xml.Element)
	if !ok {
		return fmt.Errorf("element expected")
	}
	n, err := name.ElementByName("name")
	if err != nil {
		return err
	}
	node.Value = n.Content()
	return nil
}

func getOptionsFromXml(node xml.Node, options *OptionsBag) error {
	el, ok := node.(*xml.Element)
	if !ok {
		return fmt.Errorf("element expected")
	}
	rdr, err := el.ElementByName("rendering")
	if err != nil {
		return nil
	}
	setters := makeSetters(options)
	for _, c := range rdr.Elements() {
		if c.Local == "style" {
			if err := getStyleFromXml(c, &options.Global); err != nil {
				return err
			}
		} else {
			fn, ok := setters[c.Local]
			if !ok {
				return fmt.Errorf("%s: unknown option", c.Local)
			}
			var value any
			if c.Leaf() {
				value = true
			} else {
				value = c.Content()
			}
			if err := fn(value); err != nil {
				return fmt.Errorf("%s: option(xml): %w - %s", c.Local, err, value)
			}
		}
	}
	return nil
}

func getStyleFromXml(node xml.Node, style *trellis.Style) error {
	el, ok := node.(*xml.Element)
	if !ok {
		return fmt.Errorf("element expected")
	}
	setters := makeStyleSetters(style)
	for _, c := range el.Elements() {
		fn, ok := setters[c.Local]
		if !ok {
			return fmt.Errorf("%s: unknown style property", c.Local)
		}
		var value any
		if c.Leaf() {
			value = true
		} else {
			value = c.Content()
		}
		if err := fn(value); err != nil {
			return fmt.Errorf("%s: style(xml): %w - %s", c.Local, err, value)
		}
	}
	return nil
}

func treeFromSexpr(r io.Reader) (*TreeSpec, error) {
	h := handleSexpr()
	if err := sexpr.Process(r, h, nil); err != nil {
		return nil, err
	}
	if h.root == nil {
		return nil, ErrRoot
	}
	spec := TreeSpec{
		Node:    h.root,
		Options: h.options,
	}
	return &spec, nil
}

type context struct {
	values   []any
	children []*trellis.Node
}

type sHandler struct {
	options *OptionsBag
	root    *trellis.Node
	stack   []*context

	setters map[string]func(any) error
	flags   map[string]func()
}

func handleSexpr() *sHandler {
	h := &sHandler{
		options: defaultBag(),
	}
	h.setters = makeSetters(h.options)
	h.flags = makeFlags(h.options)
	return h
}

func (h *sHandler) BeginList() error {
	var ctx context
	h.stack = append(h.stack, &ctx)
	return nil
}

func (h *sHandler) EndList() error {
	if len(h.stack) == 0 {
		return nil
	}
	var (
		ix  = len(h.stack) - 1
		ctx = h.stack[ix]
	)
	h.stack = h.stack[:ix]
	if len(ctx.values) == 0 {
		if len(h.stack) == 0 {
			return ErrRoot
		}
		parent := h.stack[len(h.stack)-1]
		parent.children = append(parent.children, ctx.children...)
		return nil
	}
	var name string
	switch str := ctx.values[0].(type) {
	case sexpr.Ident:
		name = string(str)
	case string:
		name = str
	default:
		return ErrType
	}
	n := trellis.NewNode(name)
	n.Nodes = append(n.Nodes, ctx.children...)

	if len(h.stack) == 0 {
		h.root = n
	} else {
		parent := h.stack[len(h.stack)-1]
		parent.children = append(parent.children, n)
	}

	return nil
}

func (h *sHandler) Atom(atom any) error {
	if len(h.stack) == 0 {
		return nil
	}
	x := len(h.stack) - 1
	h.stack[x].values = append(h.stack[x].values, atom)
	return nil
}

func (h *sHandler) Directive(expr []any) error {
	switch len(expr) {
	default:
		return fmt.Errorf("directive malformed")
	case 1:
		return h.handleFlag(expr[0])
	case 2:
		return h.handleOption(expr[0], expr[1])
	}
	return nil
}

func (h *sHandler) handleFlag(expr any) error {
	name, ok := expr.(sexpr.Ident)
	if !ok {
		return ErrIdent
	}
	set, ok := h.flags[string(name)]
	if !ok {
		return unknownDirective(string(name))
	}
	set()
	return nil
}

func (h *sHandler) handleOption(expr, value any) error {
	name, ok := expr.(sexpr.Ident)
	if !ok {
		return ErrIdent
	}
	setter, ok := h.setters[string(name)]
	if !ok {
		return unknownDirective(string(name))
	}
	return setter(value)
}

func unknownDirective(name string) error {
	return fmt.Errorf("%s: %w", name, ErrDirective)
}
