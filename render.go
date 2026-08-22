package trellis

import "io"

type Renderer interface {
	Render(*Tree, *Options) error
}

type Content struct {
	Value     []byte
	Bold      bool
	Italic    bool
	Underline bool
}

func HorizontalTree(w io.Writer, tree *Tree, opts *Options) error {
	r := NewHorizontal(w)
	return r.Render(tree, opts)
}

func VerticalTree(w io.Writer, tree *Tree, opts *Options) error {
	r := NewVertical(w)
	return r.Render(tree, opts)
}

func CompactTree(w io.Writer, tree *Tree, opts *Options) error {
	r := NewCompact(w)
	return r.Render(tree, opts)
}

func SunburstTree(w io.Writer, tree *Tree, opts *Options) error {
	r := NewSunburst(w)
	return r.Render(tree, opts)
}

func RadialTree(w io.Writer, tree *Tree, opts *Options) error {
	r := NewRadial(w)
	return r.Render(tree, opts)
}

type Node struct {
	Value string
	Nodes []*Node
}

func NewNode(value string) *Node {
	return &Node{
		Value: value,
	}
}

func (n *Node) Leaf() bool {
	return len(n.Nodes) == 0
}

type Tree struct {
	Root *Node
}

func NewTree(node *Node) *Tree {
	return &Tree{
		Root: node,
	}
}
