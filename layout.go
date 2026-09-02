package trellis

import (
	"slices"
	"strings"
)

type Layout interface {
	Compute(*Tree, *Options) []*Item
}

type Segment struct {
	Start Point
	End   Point
}

func (s Segment) Swap() Segment {
	s.Start, s.End = s.End, s.Start
	return s
}

type Point struct {
	X, Y int
}

func (p Point) Swap() Point {
	p.X, p.Y = p.Y, p.X
	return p
}

func (p Point) IsAbove(other Point) bool {
	return p.Y <= other.Y
}

func (p Point) IsLeft(other Point) bool {
	return p.X <= other.X
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

func horizontalPath(from, to *Item) []Segment {
	if !from.Point.IsLeft(to.Point) {
		from, to = to, from
	}
	var (
		start  = from.Point
		end    = to.Point
		offset = len(from.Value)
	)
	if start.Y == end.Y {
		s := Segment{
			Start: start,
			End:   end,
		}
		s.Start.X += offset
		s.End.X--
		return []Segment{s}
	}
	f := Segment{
		Start: start,
		End:   start,
	}
	f.Start.X += offset
	f.End.X = from.W.End

	t := Segment{
		Start: end,
		End:   end,
	}
	t.Start.X = to.W.Start
	t.End.X--

	var v Segment
	if f.End.IsAbove(t.Start) {
		v.Start, v.End = f.End, t.Start
	} else {
		v.Start, v.End = t.Start, f.End
	}

	return []Segment{f, v, t}
}

func verticalPath(from, to *Item) []Segment {
	if !from.Point.IsAbove(to.Point) {
		from, to = to, from
	}
	var (
		start = from.Point
		end   = to.Point
	)
	start.X += from.Size() / 2
	end.X += to.Size() / 2

	if start.X == end.X {
		s := Segment{
			Start: start,
			End:   end,
		}
		s.Start.Y++
		s.End.Y--
		return []Segment{s.Swap()}
	}
	f := Segment{
		Start: start,
		End:   start,
	}
	f.Start.Y++
	f.End.Y = from.H.End

	t := Segment{
		Start: end,
		End:   end,
	}
	t.Start.Y = to.H.Start
	t.End.Y--

	var v Segment
	if f.End.IsLeft(t.Start) {
		v.Start, v.End = f.End, t.Start
	} else {
		v.Start, v.End = t.Start, f.End
	}
	return []Segment{f.Swap(), v, t.Swap()}
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
		i.X = i.W.End - len(i.Value)
	default:
		i.X = i.W.Start + i.W.Offset() - (len(i.Value) / 2)
	}
}

func (i *Item) MoveX(delta int) {
	i.X += delta
	i.W.Start += delta
	i.W.End += delta

	for _, c := range i.Children {
		c.MoveX(delta)
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

func (i *Item) Size() int {
	return len(i.Value)
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
	var items []*Item
	switch opts.Orient {
	case HorizontalLayout:
		items = i.horizontal(tree, opts)
	default:
		items = i.vertical(tree, opts)
	}
	opts.Width = maxFromItems(items, func(i *Item) int { return i.W.End })
	opts.Height = maxFromItems(items, func(i *Item) int { return i.H.End })
	return items
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
	)
	ix := slices.IndexFunc(is, func(it *Item) bool {
		return it.Root()
	})
	if ix < 0 {
		return nil
	}
	i.computeVerticalCoordinates(is[ix], opts, spacing, level)

	return is
}

func (i ideal) computeVerticalCoordinates(node *Item, opts *Options, spacing, level int) {
	var (
		height = opts.Height / (level + 1)
		maxLen int
	)
	for _, x := range node.Children {
		if !x.Leaf() {
			i.computeVerticalCoordinates(x, opts, spacing, level)
			continue
		}
		var (
			startY = (x.Default.Y * height)
			startX = (x.Default.X * opts.Width / spacing)
			endX   = ((x.Default.X + opts.SiblingGap) * opts.Width) / spacing
		)
		x.X = startX
		x.Y = startY
		x.W = NewSpan(startX, endX)
		x.H = NewSpan(startY, startY+height)

		if x.W.Len() < opts.SiblingGap+1 {
			x.W.End = x.W.Start + opts.SiblingGap + 1
		}
		x.AlignX(opts.AlignX)
		x.AlignY(opts.AlignY)

		maxLen = max(maxLen, x.W.Len())
	}

	boundary := node.Children[0].W.End
	for _, c := range node.Children[1:] {
		if c.W.Start < boundary {
			c.MoveX(boundary - c.W.Start + 1)
		}
		boundary = c.W.End
	}
	var (
		first = node.FirstLeaf()
		last  = node.LastLeaf()
	)
	node.X = node.Default.X * opts.Width / spacing
	node.Y = node.Default.Y * height
	node.W = NewSpan(first.W.Start, last.W.End)
	node.H = NewSpan(node.Y, node.Y+height)
	node.AlignX(opts.AlignX)
	node.AlignY(opts.AlignY)
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
		if opts.Align() == AlignStart {
			sub.Y = sub.Children[0].Y
		} else if opts.Align() == AlignEnd {
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
