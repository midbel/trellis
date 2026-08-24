package trellis

type span struct {
	Start int
	End   int
}

func createSpan(start, end int) span {
	return span{
		Start: start,
		End:   end,
	}
}

func (s span) CenterValue(value string, padding int) int {
	var (
		size   = len([]byte(value)) + (2 * (padding))
		offset = (s.Len() - size) / 2
	)
	return s.Start + offset
}

func (s span) Distance(other span) int {
	return other.Center() - s.Center()
}

func (s span) Center() int {
	return s.Start + s.Offset()
}

func (s span) Offset() int {
	return s.Len() / 2
}

func (s span) Len() int {
	return s.End - s.Start
}

func (s span) Next() span {
	sp := span{
		Start: s.End + 1,
		End:   s.End + 1 + s.Len(),
	}
	return sp
}

type layoutNode struct {
	*Node
	root bool
	X    int
	Y    int

	W span
	H span

	Children []*layoutNode
}

func (n *layoutNode) Leaf() bool {
	return len(n.Children) == 0
}

func (n *layoutNode) Root() bool {
	return n.root
}

func (n *layoutNode) Len() int {
	return len(n.Children)
}

func (n *layoutNode) Weight() int {
	if n.Leaf() {
		return 1
	}
	var sum int
	for _, x := range n.Children {
		sum += x.Weight()
	}
	return sum
}

func (n *layoutNode) Anchor(padding int) int {
	x := n.Get(padding)
	return n.X + len(x)/2
}

func (n *layoutNode) Get(padding int) []byte {
	var (
		value = []byte(n.Value)
		size  = len(value)
	)
	if padding > 0 {
		size += 2 * padding
		tmp := make([]byte, size)
		for i := range tmp {
			tmp[i] = ' '
		}
		copy(tmp[padding:], value)
		value = tmp
	}
	return value
}

type Point struct {
	X, Y int
}

type Coordinate struct {
	*Node
	Ideal    Point
	Computed Point
	Width    int
	Height   int
}

func ComputeLayout(tree *Tree, options *Options) []Coordinate {
	var (
		opts   = prepareOptions(options)
		maker  = makeLayout(opts.SiblingGap, opts.Align)
		layout = maker.Make(tree.Root)
	)
	var nodes []Coordinate
	for _, x := range layout {
		c := Coordinate{
			Node: x.Node,
			Ideal: Point{
				X: x.X,
				Y: x.Y,
			},
			Computed: Point{
				X: x.X,
				Y: x.Y,
			},
		}
		nodes = append(nodes, c)
	}
	return nodes
}

type layoutMaker struct {
	nextLeafPosition int
	gapSize          int
	maxDepth         int
	align            Alignment
}

func makeLayout(gap int, align Alignment) *layoutMaker {
	return &layoutMaker{
		gapSize: gap,
		align:   align,
	}
}

func (m *layoutMaker) Single(node *Node) *layoutNode {
	return m.makeLayout(node, 0)
}

func (m *layoutMaker) Make(node *Node) []*layoutNode {
	res := m.makeLayout(node, 0)
	return flatten(res)
}

func (m *layoutMaker) Depth() int {
	return m.maxDepth + 1
}

func (m *layoutMaker) Spacing() int {
	return m.nextLeafPosition
}

func (m *layoutMaker) makeLayout(node *Node, depth int) *layoutNode {
	sub := layoutNode{
		Node: node,
		X:    depth,
		root: depth == 0,
	}
	depth++
	for _, c := range node.Nodes {
		child := m.makeLayout(c, depth)
		sub.Children = append(sub.Children, child)
	}
	if node.Leaf() {
		sub.Y = m.nextLeafPosition
		m.nextLeafPosition += m.gapSize
	} else {
		if m.align == AlignStart {
			sub.Y = sub.Children[0].Y
		} else if m.align == AlignEnd {
			sub.Y = sub.Children[len(sub.Children)-1].Y
		} else {
			var sum int
			for i := range sub.Children {
				sum += sub.Children[i].Y
			}
			sub.Y = sum / (len(sub.Children))
		}
	}
	m.maxDepth = max(depth-1, m.maxDepth)
	return &sub
}

func flatten(node *layoutNode) []*layoutNode {
	list := []*layoutNode{
		node,
	}
	for _, c := range node.Children {
		list = append(list, flatten(c)...)
	}
	return list
}

func getOffsetX(align Alignment, width, size int) int {
	switch align {
	case AlignStart:
		return 0
	case AlignBottom:
		return width - size
	default:
		return (width - size) / 2
	}
}
