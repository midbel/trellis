package codec

import (
	"errors"
	"fmt"
	"io"

	"github.com/midbel/sexpr"
	"github.com/midbel/trellis"
)

var (
	ErrFormat    = errors.New("format not supported")
	ErrIdent     = errors.New("identifier expected")
	ErrDirective = errors.New("unsupported directive")
	ErrType      = errors.New("invalid value type")
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
	Type string
	*trellis.Tree
	*trellis.Options
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

func treeFromXml(r io.Reader) (*TreeSpec, error) {
	return nil, nil
}

func treeFromJson(r io.Reader) (*TreeSpec, error) {
	return nil, nil
}

func treeFromSexpr(r io.Reader) (*TreeSpec, error) {
	var h handler
	h.options = new(trellis.Options)
	if err := sexpr.Process(r, &h); err != nil {
		return nil, err
	}
	spec := TreeSpec{
		Type:    h.Type,
		Tree:    trellis.NewTree(h.root),
		Options: h.options,
	}
	return &spec, nil
}

type context struct {
	values   []any
	children []*trellis.Node
}

type handler struct {
	Type    string
	options *trellis.Options
	root    *trellis.Node
	stack   []*context
}

func (h *handler) BeginList() error {
	var ctx context
	h.stack = append(h.stack, &ctx)
	return nil
}

func (h *handler) EndList() error {
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
			return nil
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
		name = "unknown"
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

func (h *handler) Atom(atom any) error {
	if len(h.stack) == 0 {
		return nil
	}
	x := len(h.stack) - 1
	h.stack[x].values = append(h.stack[x].values, atom)
	return nil
}

func (h *handler) Directive(expr []any) error {
	switch len(expr) {
	default:
	case 1:
		return h.handleFlag(expr[0])
	case 2:
		return h.handleOption(expr[0], expr[1])
	}
	return nil
}

func (h *handler) handleFlag(expr any) error {
	name, ok := expr.(sexpr.Ident)
	if !ok {
		return ErrIdent
	}
	switch name {
	default:
		return unknownDirective(string(name))
	case "border":
		h.options.Border = true
	case "coordinates":
		h.options.ShowCoordinates = true
	case "reverse":
		h.options.Reverse = true
	case "padding":
		h.options.Padding++
	case "horizontalGap":
		h.options.HorizontalGap++
	case "verticalGap":
		h.options.VerticalGap++
	case "horizontal", "vertical", "compact":
		h.Type = string(name)
	}
	return nil
}

func (h *handler) handleOption(expr, value any) error {
	name, ok := expr.(sexpr.Ident)
	if !ok {
		return ErrIdent
	}
	var err error
	switch name {
	case "type":
		err = setString(&h.Type, value)
	case "width":
		err = setInt(&h.options.Width, value)
	case "height":
		err = setInt(&h.options.Height, value)
	case "padding":
		err = setInt(&h.options.Padding, value)
	case "paddingChar":
		err = setString(&h.options.PaddingChar, value)
	case "horizontalGap":
		err = setInt(&h.options.HorizontalGap, value)
	case "verticalGap":
		err = setInt(&h.options.VerticalGap, value)
	case "border":
		err = setBool(&h.options.Border, value)
	case "coordinates":
		err = setBool(&h.options.ShowCoordinates, value)
	case "reverse":
		err = setBool(&h.options.Reverse, value)
	default:
		return unknownDirective(string(name))
	}
	return err
}

func setBool(b *bool, value any) error {
	x, ok := value.(bool)
	if !ok {
		return ErrType
	}
	*b = x
	return nil
}

func setString(s *string, value any) error {
	switch v := value.(type) {
	case string:
		*s = v
	case sexpr.Ident:
		*s = string(v)
	default:
		return ErrType
	}
	return nil
}

func setInt(i *int, value any) error {
	switch v := value.(type) {
	case int:
		*i = v
	case int8:
		*i = int(v)
	case int16:
		*i = int(v)
	case int32:
		*i = int(v)
	case int64:
		*i = int(v)
	default:
		return ErrType
	}
	return nil
}

func unknownDirective(name string) error {
	return fmt.Errorf("%s: %w", name, ErrDirective)
}
