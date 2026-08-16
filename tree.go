package trellis

import "io"

type Renderer interface {
	Render(*Tree, *TreeRenderOptions) error
}

func HorizontalTree(w io.Writer, tree *Tree, opts *TreeRenderOptions) error {
	r := NewHorizontal(w)
	return r.Render(tree, opts)
}

func VerticalTree(w io.Writer, tree *Tree, opts *TreeRenderOptions) error {
	r := NewVertical(w)
	return r.Render(tree, opts)
}

func CompactTree(w io.Writer, tree *Tree, opts *TreeRenderOptions) error {
	r := NewCompact(w)
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
