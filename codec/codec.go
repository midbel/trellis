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
			return err
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
	if b, ok := y["bold"].(bool); ok && b {
		style.Flags |= trellis.StyleBold
	}
	if i, ok := y["italic"].(bool); ok && i {
		style.Flags |= trellis.StyleItalic
	}
	if u, ok := y["underline"].(bool); ok && u {
		style.Flags |= trellis.StyleUnderline
	}
	if f, ok := y["font-family"].(string); ok && f != "" {
		style.Family = f
	}
	if a, ok := y["align-x"].(string); ok {
		x, err := parseAlignment(a)
		if err != nil {
			return err
		}
		style.AlignX = x
	}
	if a, ok := y["align-y"].(string); ok {
		x, err := parseAlignment(a)
		if err != nil {
			return err
		}
		style.AlignY = x
	}
	if v, ok := y["font-size"]; ok {
		z, err := parseNumber(v)
		if err != nil {
			return err
		}
		style.Size = z
	}
	if v, ok := y["margin"]; ok {
		z, err := parseInsets(v)
		if err != nil {
			return err
		}
		style.Margin = trellis.NewInsets(z, z, z, z)
	}
	if v, ok := y["padding"]; ok {
		z, err := parseInsets(v)
		if err != nil {
			return err
		}
		style.Padding = trellis.NewInsets(z, z, z, z)
	}
	if c, ok := y["color"].(string); ok && c != "" {
		style.Color = c
	}
	return nil
}

func treeFromXml(r io.Reader) (*TreeSpec, error) {
	var (
		rs = xml.NewReader(r)
		xh = handleXml()
	)
	if err := rs.Read(xh); err != nil {
		return nil, err
	}
	return nil, nil
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

type xHandler struct {
	options *OptionsBag
	root    *trellis.Node
	stack   []*context

	setters map[string]func(any) error
	flags   map[string]func()
}

func handleXml() *xHandler {
	h := &xHandler{
		options: defaultBag(),
	}
	h.setters = makeSetters(h.options)
	h.flags = makeFlags(h.options)
	return h
}

func (h *xHandler) OnStartElement(el xml.Element, self bool) error {
	return nil
}

func (h *xHandler) OnCloseElement(name xml.Name) error {
	return nil
}

func (h *xHandler) OnText(txt xml.Text) error {
	return nil
}

func (h *xHandler) OnComment(comment xml.Comment) error {
	return nil
}

func (h *xHandler) OnPI(pi xml.PI) error {
	return nil
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
