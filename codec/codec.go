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
	Type string
	*trellis.Node
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
	h := newHandler()
	if err := sexpr.Process(r, h, nil); err != nil {
		return nil, err
	}
	if h.root == nil {
		return nil, ErrRoot
	}
	spec := TreeSpec{
		Type:    h.Type,
		Node:    h.root,
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

	setters map[string]func(any) error
	flags   map[string]func()
}

func newHandler() *handler {
	h := &handler{
		options: new(trellis.Options),
	}
	h.setters = makeSetters(h.options)
	h.flags = makeFlags(h.options)
	return h
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
	if name == "vertical" || name == "horizontal" || name == "compact" {
		h.Type = string(name)
		return nil
	}
	set, ok := h.flags[string(name)]
	if !ok {
		return unknownDirective(string(name))
	}
	set()
	return nil
}

func (h *handler) handleOption(expr, value any) error {
	name, ok := expr.(sexpr.Ident)
	if !ok {
		return ErrIdent
	}
	if name == "type" {
		str, err := parseString(value)
		if err == nil {
			h.Type = str
		}
		return err
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
