package trellis

import "io"

type Renderer interface {
	Render(*Node, *Options) error
}

func HorizontalTree(w io.Writer, root *Node, opts *Options) error {
	r := NewHorizontal(w)
	return r.Render(root, opts)
}

func VerticalTree(w io.Writer, root *Node, opts *Options) error {
	r := NewVertical(w)
	return r.Render(root, opts)
}

func CompactTree(w io.Writer, root *Node, opts *Options) error {
	r := NewCompact(w)
	return r.Render(root, opts)
}

func SunburstTree(w io.Writer, root *Node, opts *Options) error {
	r := NewSunburst(w)
	return r.Render(root, opts)
}

func RadialTree(w io.Writer, root *Node, opts *Options) error {
	r := NewRadial(w)
	return r.Render(root, opts)
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
