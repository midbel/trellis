package trellis

import (
	"slices"
	"strings"
)

type Layout interface {
	Compute(*Tree, *Options) []*Item
}

type Point struct {
	X, Y int
}

func (p Point) Swap() Point {
	p.X, p.Y = p.Y, p.X
	return p
}

type CoordinateMap struct {
	Width       int
	Height      int
	Coordinates []Coordinate
}

type Coordinate struct {
	Value    string
	Ideal    Point
	Computed Point
	Width    Span
	Height   Span
}

func ComputeLayout(tree *Tree, options *Options) CoordinateMap {
	var (
		opts = prepareOptions(options)
		mk   = Ideal()
		is   = mk.Compute(tree, opts)
		res  CoordinateMap
	)
	for i := range is {
		c := Coordinate{
			Value: strings.TrimSpace(is[i].String()),
			Ideal: Point{
				X: is[i].Default.X,
				Y: is[i].Default.Y,
			},
			Computed: Point{
				X: is[i].X,
				Y: is[i].Y,
			},
			Width:  is[i].W,
			Height: is[i].H,
		}
		res.Coordinates = append(res.Coordinates, c)
	}
	res.Width = opts.Width
	res.Height = opts.Height
	return res
}

type Item struct {
	Content

	Default Point
	Point
	W Span
	H Span

	Children []*Item
	root     bool
}

func maxFromItems(is []*Item, get func(*Item) int) int {
	var res int
	for i := range is {
		res = max(get(is[i]), res)
	}
	return res
}

func (i *Item) FirstLeaf() *Item {
	if i.Leaf() {
		return i
	}
	return i.Children[0].FirstLeaf()
}

func (i *Item) LastLeaf() *Item {
	if i.Leaf() {
		return i
	}
	n := i.Len() - 1
	return i.Children[n].LastLeaf()
}

func (i *Item) AlignY(align Alignment) {
	switch align {
	case AlignStart:
		i.Y = i.H.Start
	case AlignEnd:
		i.Y = i.H.End
	default:
		i.Y = i.H.Start + i.H.Offset()
	}
}

func (i *Item) AlignX(align Alignment) {
	switch align {
	case AlignStart:
		i.X = i.W.Start
	case AlignEnd:
		i.X = i.W.End
	default:
		i.X = i.W.Start + i.W.Offset()
	}
}

func (i *Item) MoveY(delta int) {
	i.Y += delta
	i.H.Start += delta
	i.H.End += delta

	for _, c := range i.Children {
		c.MoveY(delta)
	}
}

func (i *Item) Leaf() bool {
	return i.Len() == 0
}

func (i *Item) Root() bool {
	return i.root
}

func (i *Item) Len() int {
	return len(i.Children)
}

func (i *Item) Weight() int {
	if i.Leaf() {
		return 1
	}
	var sum int
	for _, x := range i.Children {
		sum += x.Weight()
	}
	return sum
}

func Ideal() Layout {
	var i ideal
	return i
}

func Proportional() Layout {
	var p proportional
	return p
}

type ideal struct{}

func (i ideal) Compute(tree *Tree, opts *Options) []*Item {
	switch opts.Orient {
	case HorizontalLayout:
		return i.horizontal(tree, opts)
	default:
		return i.vertical(tree, opts)
	}
}

func (i ideal) vertical(tree *Tree, opts *Options) []*Item {
	is := i.prepare(tree, opts)
	for i := range is {
		is[i].Point = is[i].Swap()
		is[i].Default = is[i].Point
	}
	var (
		spacing = maxFromItems(is, func(i *Item) int { return i.X })
		level   = maxFromItems(is, func(i *Item) int { return i.Y })
		width   = opts.Width / spacing
		height  = opts.Height / (level + 1)
	)
	for i := range is {
		var (
			startX = (width * is[i].Default.X) + opts.borderWidth()
			startY = (height * is[i].Default.Y) + opts.borderWidth()
		)
		is[i].W = NewSpan(startX, startX+width-opts.borderWidth())
		is[i].H = NewSpan(startY, startY+height-opts.borderWidth())

		is[i].Y = is[i].H.Start
		is[i].X = startX
	}
	return is
}

func (i ideal) horizontal(tree *Tree, opts *Options) []*Item {
	var (
		is      = i.prepare(tree, opts)
		spacing = maxFromItems(is, func(i *Item) int { return i.Y })
		level   = maxFromItems(is, func(i *Item) int { return i.X })
	)
	ix := slices.IndexFunc(is, func(it *Item) bool {
		return it.Root()
	})
	if ix < 0 {
		return nil
	}
	i.computeHorizontalCoordinates(is[ix], opts, spacing, level)
	return is
}

func (i ideal) computeHorizontalCoordinates(node *Item, opts *Options, spacing, level int) {
	var (
		width  = opts.Width / (level + 1)
		maxLen int
	)
	for _, x := range node.Children {
		if !x.Leaf() {
			i.computeHorizontalCoordinates(x, opts, spacing, level)
			continue
		}
		var (
			startX = (x.Default.X * width)
			startY = (x.Default.Y * opts.Height / spacing)
			endY   = ((x.Default.Y + opts.SiblingGap) * opts.Height) / spacing
		)
		x.X = startX
		x.Y = startY
		x.W = NewSpan(startX, startX+width)
		x.H = NewSpan(startY, endY)

		if x.H.Len() < opts.SiblingGap+1 {
			x.H.End = x.H.Start + opts.SiblingGap + 1
		}
		x.AlignX(opts.AlignX)
		x.AlignY(opts.AlignY)

		maxLen = max(maxLen, x.H.Len())
	}

	boundary := node.Children[0].H.End
	for _, c := range node.Children[1:] {
		if c.H.Start < boundary {
			c.MoveY(boundary - c.H.Start + 1)
		}
		boundary = c.H.End
	}
	var (
		first = node.FirstLeaf()
		last  = node.LastLeaf()
	)
	node.X = node.Default.X * width
	node.Y = node.Default.Y * opts.Height / spacing
	node.W = NewSpan(node.X, node.X+width)
	node.H = NewSpan(first.H.Start, last.H.End)
	node.AlignX(opts.AlignX)
	node.AlignY(opts.AlignY)
}

func (ideal) prepare(tree *Tree, opts *Options) []*Item {
	var (
		mk = defaultTreeLayout()
		is = mk.Make(tree.Root, opts)
	)
	if opts.Reverse {
		level := mk.Depth() - 1
		for i := range is {
			is[i].Default.X = level - is[i].X
			is[i].Point = is[i].Default
		}
	}
	return is
}

type proportional struct{}

func (p proportional) Compute(tree *Tree, opts *Options) []*Item {
	switch opts.Orient {
	case HorizontalLayout:
		return p.horizontal(tree, opts)
	default:
		return p.vertical(tree, opts)
	}
}

func (proportional) vertical(tree *Tree, opts *Options) []*Item {
	return nil
}

func (proportional) horizontal(tree *Tree, opts *Options) []*Item {
	return nil
}

type Span struct {
	Start int
	End   int
}

func NewSpan(start, end int) Span {
	return Span{
		Start: start,
		End:   end,
	}
}

func (s Span) CenterValue(value string, padding int) int {
	var (
		size   = len([]byte(value)) + (2 * (padding))
		offset = (s.Len() - size) / 2
	)
	return s.Start + offset
}

func (s Span) Offset() int {
	return s.Len() / 2
}

func (s Span) Len() int {
	return s.End - s.Start
}

func (s Span) Next() Span {
	return NewSpan(s.End+1, s.End+1+s.Len())
}

type treeLayout struct {
	siblingsSpacing int
	levelSpacing    int
}

func defaultTreeLayout() *treeLayout {
	return &treeLayout{}
}

func (m *treeLayout) Single(node *Node, opts *Options) *Item {
	return m.makeLayout(node, 0, opts)
}

func (m *treeLayout) Make(node *Node, opts *Options) []*Item {
	res := m.makeLayout(node, 0, opts)
	return m.flatten(res)
}

func (m *treeLayout) Depth() int {
	return m.levelSpacing + 1
}

func (m *treeLayout) Spacing() int {
	return m.siblingsSpacing
}

func (m *treeLayout) makeLayout(node *Node, depth int, opts *Options) *Item {
	sub := Item{
		Content: opts.Render(node, opts),
		root:    depth == 0,
	}
	sub.X = depth
	depth++
	for _, n := range node.Nodes {
		child := m.makeLayout(n, depth, opts)
		sub.Children = append(sub.Children, child)
	}
	if node.Leaf() {
		sub.Y = m.siblingsSpacing
		m.siblingsSpacing += opts.SiblingGap
	} else {
		if opts.Align == AlignStart {
			sub.Y = sub.Children[0].Y
		} else if opts.Align == AlignEnd {
			sub.Y = sub.Children[len(sub.Children)-1].Y
		} else {
			var sum int
			for i := range sub.Children {
				sum += sub.Children[i].Y
			}
			sub.Y = sum / (len(sub.Children))
		}
	}
	sub.Default = sub.Point
	m.levelSpacing = max(depth-1, m.levelSpacing)
	return &sub
}

func (m *treeLayout) flatten(node *Item) []*Item {
	list := []*Item{
		node,
	}
	for _, n := range node.Children {
		list = append(list, m.flatten(n)...)
	}
	return list
}

type layoutNode struct {
	*Node
	root bool
	X    int
	Y    int

	W Span
	H Span

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
