package trellis

import (
	"slices"
	"strings"
)

type Layout interface {
	Compute(*Node, *Options) []*Item
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

func ComputeLayout(root *Node, options *Options) (CoordinateMap, error) {
	opts, err := prepareOptions(options)
	if err != nil {
		return CoordinateMap{}, err
	}
	var (
		mk  = Ideal()
		is  = mk.Compute(root, opts)
		res CoordinateMap
	)
	for i := range is {
		c := Coordinate{
			Value: strings.TrimSpace(is[i].String()),
			Ideal: Point{
				X: is[i].Ideal.X,
				Y: is[i].Ideal.Y,
			},
			Computed: Point{
				X: is[i].Position.X,
				Y: is[i].Position.Y,
			},
			Width:  is[i].W,
			Height: is[i].H,
		}
		res.Coordinates = append(res.Coordinates, c)
	}
	res.Width = opts.Width
	res.Height = opts.Height
	return res, nil
}

type Segment struct {
	Start Point
	End   Point
}

func (s Segment) DistanceX() int {
	return s.End.X - s.Start.X
}

func (s Segment) DistanceY() int {
	return s.End.Y - s.Start.Y
}

func (s Segment) One(other Segment) bool {
	return s.Start.Equal(other.Start) && s.End.Equal(other.End)
}

func (s Segment) Swap() Segment {
	s.Start, s.End = s.End, s.Start
	return s
}

func (s Segment) Horizontal() bool {
	return s.Start.Y == s.End.Y
}

func (s Segment) Vertical() bool {
	return s.Start.X == s.End.X
}

type Point struct {
	X, Y int
}

func (p Point) Equal(other Point) bool {
	return p.X == other.X && p.Y == other.Y
}

func (p Point) Swap() Point {
	p.X, p.Y = p.Y, p.X
	return p
}

func (p Point) BeforeY(other Point) bool {
	return p.Y <= other.Y
}

func (p Point) BeforeX(other Point) bool {
	return p.X < other.X
}

func horizontalPath(from, to *Item, opts *Options) []Segment {
	if !from.Position.BeforeX(to.Position) {
		from, to = to, from
	}
	var (
		start  = from.Position
		end    = to.Position
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
	f.End.X = from.W.End + opts.Margin

	t := Segment{
		Start: end,
		End:   end,
	}
	t.Start.X = to.W.Start - opts.Margin
	t.End.X--

	var v Segment
	if f.End.BeforeY(t.Start) {
		v.Start, v.End = f.End, t.Start
	} else {
		v.Start, v.End = t.Start, f.End
	}

	return []Segment{f, v, t}
}

func verticalPath(from, to *Item, opts *Options) []Segment {
	if !from.Position.BeforeY(to.Position) {
		from, to = to, from
	}
	var (
		start = from.Position
		end   = to.Position
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
		return []Segment{s}
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
	if f.End.BeforeX(t.Start) {
		v.Start, v.End = f.End, t.Start
	} else {
		v.Start, v.End = t.Start, f.End
	}
	return []Segment{f, v, t}
}

type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (r Rect) StartX() int {
	return r.X
}

func (r Rect) EndX() int {
	return r.X + r.Width
}

func (r Rect) StartY() int {
	return r.Y
}

func (r Rect) EndY() int {
	return r.Y + r.Height
}

type Item struct {
	Content

	Ideal    Point
	Position Point
	Bounds   Rect
	W        Span
	H        Span

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
		i.Position.Y = i.H.Start
	case AlignEnd:
		i.Position.Y = i.H.End
	default:
		i.Position.Y = i.H.Start + i.H.Offset()
	}
}

func (i *Item) AlignX(align Alignment) {
	switch align {
	case AlignStart:
		i.Position.X = i.W.Start
	case AlignEnd:
		i.Position.X = i.W.End - len(i.Value)
	default:
		i.Position.X = i.W.Start + i.W.Offset() - (len(i.Value) / 2)
	}
}

func (i *Item) MoveX(delta int) {
	i.Position.X += delta
	i.W.Start += delta
	i.W.End += delta

	for _, c := range i.Children {
		c.MoveX(delta)
	}
}

func (i *Item) MoveY(delta int) {
	i.Position.Y += delta
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

func (i ideal) Compute(root *Node, opts *Options) []*Item {
	var items []*Item
	switch opts.Orient {
	case HorizontalLayout:
		items = i.horizontalLayout(root, opts)
	case VerticalLayout:
		items = i.verticalLayout(root, opts)
	default:
		return i.compactLayout(root, opts)
	}
	opts.Width = maxFromItems(items, func(i *Item) int { return i.W.End })
	opts.Height = maxFromItems(items, func(i *Item) int { return i.H.End })
	return items
}

func (i ideal) compactLayout(root *Node, opts *Options) []*Item {
	spacing := opts.Spacing
	opts.Spacing = 1

	items := i.prepare(root, opts)
	for i := 1; i < len(items); i++ {
		for items[i].Position.Y <= items[i-1].Position.Y {
			items[i].Position.Y++
		}
		items[i].Position.X += spacing
		opts.Height = items[i].Position.Y
	}
	opts.Height++

	ix := slices.IndexFunc(items, func(i *Item) bool {
		return i.Root()
	})
	rearrangeCompactChildren(items[ix], spacing)
	return items
}

func rearrangeCompactChildren(node *Item, spacing int) {
	for i := range node.Children {
		node.Children[i].Position.X = node.Position.X + spacing
		rearrangeCompactChildren(node.Children[i], spacing)
	}
}

func (i ideal) verticalLayout(root *Node, opts *Options) []*Item {
	is := i.prepare(root, opts)
	for i := range is {
		is[i].Position = is[i].Position.Swap()
		is[i].Ideal = is[i].Position
	}
	var (
		spacing = maxFromItems(is, func(i *Item) int { return i.Position.X })
		level   = maxFromItems(is, func(i *Item) int { return i.Position.Y })
	)
	ix := slices.IndexFunc(is, func(it *Item) bool {
		return it.Root()
	})
	if ix < 0 {
		return nil
	}
	computeVerticalCoordinates(is[ix], opts, spacing, level)
	return is
}

func computeVerticalCoordinates(node *Item, opts *Options, spacing, level int) {
	height := opts.Height / (level + 1)

	computeVerticalChildren(node, opts, spacing, level, height)
	resolveVerticalChildren(node)
	computeVerticalNode(node, opts, spacing, height)
}

func computeVerticalChildren(node *Item, opts *Options, spacing, level, height int) {
	for _, x := range node.Children {
		if !x.Leaf() {
			computeVerticalCoordinates(x, opts, spacing, level)
			continue
		}
		var (
			startY = (x.Ideal.Y * height)
			startX = (x.Ideal.X * opts.Width / spacing)
			endX   = ((x.Ideal.X + opts.Spacing) * opts.Width) / spacing
		)
		x.Position.X = startX
		x.Position.Y = startY
		x.W = NewSpan(startX, endX)
		x.H = NewSpan(startY, startY+height)

		if x.W.Len() < opts.Spacing+1 {
			x.W.End = x.W.Start + opts.Spacing + 1
		}
		x.AlignX(opts.AlignX)
		x.AlignY(opts.AlignY)
	}
}

func resolveVerticalChildren(node *Item) {
	boundary := node.Children[0].W.End
	for _, c := range node.Children[1:] {
		if c.W.Start < boundary {
			c.MoveX(boundary - c.W.Start + 1)
		}
		boundary = c.W.End
	}
}

func computeVerticalNode(node *Item, opts *Options, spacing, height int) {
	var (
		first = node.FirstLeaf()
		last  = node.LastLeaf()
	)
	node.Position.X = node.Ideal.X * opts.Width / spacing
	node.Position.Y = node.Ideal.Y * height
	node.W = NewSpan(first.W.Start, last.W.End)
	node.H = NewSpan(node.Position.Y, node.Position.Y+height)
	node.AlignX(opts.AlignX)
	node.AlignY(opts.AlignY)
}

func (i ideal) horizontalLayout(root *Node, opts *Options) []*Item {
	var (
		is      = i.prepare(root, opts)
		spacing = maxFromItems(is, func(i *Item) int { return i.Position.Y })
		level   = maxFromItems(is, func(i *Item) int { return i.Position.X })
	)
	ix := slices.IndexFunc(is, func(it *Item) bool {
		return it.Root()
	})
	if ix < 0 {
		return nil
	}
	computeHorizontalCoordinates(is[ix], opts, spacing, level)
	return is
}

func computeHorizontalCoordinates(node *Item, opts *Options, spacing, level int) {
	width := opts.Width / (level + 1)

	computeHorizontalChildren(node, opts, spacing, level, width)
	resolveHorizontalChildren(node, opts)
	computeHorizontalNode(node, opts, spacing, width)
}

func computeHorizontalChildren(node *Item, opts *Options, spacing, level, width int) {
	for _, x := range node.Children {
		if !x.Leaf() {
			computeHorizontalCoordinates(x, opts, spacing, level)
			continue
		}
		var (
			startX = x.Ideal.X * width
			startY = (x.Ideal.Y * opts.Height / spacing)
			endY   = ((x.Ideal.Y + opts.Spacing) * opts.Height) / spacing
		)
		x.Position.X = startX + opts.Margin
		x.Position.Y = startY + opts.Margin
		x.W = NewSpan(startX+opts.Margin, startX+width-opts.Margin)
		x.H = NewSpan(startY+opts.Margin, endY-opts.Margin)

		if x.H.Len() < opts.Spacing+1 {
			x.H.End = x.H.Start + opts.Spacing + 1
		}
		x.AlignX(opts.AlignX)
		x.AlignY(opts.AlignY)
	}
}

func resolveHorizontalChildren(node *Item, opts *Options) {
	boundary := node.Children[0].H.End + opts.Margin
	for _, c := range node.Children[1:] {
		if c.H.Start < boundary {
			c.MoveY(boundary - c.H.Start - opts.Margin + 1)
		}
		boundary = c.H.End + opts.Margin
	}
}

func computeHorizontalNode(node *Item, opts *Options, spacing, width int) {
	var (
		first = node.FirstLeaf()
		last  = node.LastLeaf()
	)
	node.Position.X = node.Ideal.X * width
	node.Position.Y = node.Ideal.Y * opts.Height / spacing
	node.W = NewSpan(node.Position.X+opts.Margin, node.Position.X+width-opts.Margin)
	node.Position.X += opts.Margin
	node.H = NewSpan(first.H.Start, last.H.End)
	node.AlignX(opts.AlignX)
	node.AlignY(opts.AlignY)
}

func (ideal) prepare(root *Node, opts *Options) []*Item {
	var (
		mk = defaultTreeLayout()
		is = mk.Make(root, opts)
	)
	if opts.Reverse {
		level := mk.Depth() - 1
		for i := range is {
			is[i].Ideal.X = level - is[i].Position.X
			is[i].Position = is[i].Ideal
		}
	}
	return is
}

type proportional struct{}

func (p proportional) Compute(root *Node, opts *Options) []*Item {
	switch opts.Orient {
	case HorizontalLayout:
		return p.horizontal(root, opts)
	default:
		return p.vertical(root, opts)
	}
}

func (proportional) vertical(root *Node, opts *Options) []*Item {
	return nil
}

func (proportional) horizontal(root *Node, opts *Options) []*Item {
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
	sub.Position.X = depth
	depth++
	for _, n := range node.Nodes {
		child := m.makeLayout(n, depth, opts)
		sub.Children = append(sub.Children, child)
	}
	if node.Leaf() {
		sub.Position.Y = m.siblingsSpacing
		m.siblingsSpacing += opts.Spacing
	} else {
		if opts.Align() == AlignStart {
			sub.Position.Y = sub.Children[0].Position.Y
		} else if opts.Align() == AlignEnd {
			sub.Position.Y = sub.Children[len(sub.Children)-1].Position.Y
		} else {
			var sum int
			for i := range sub.Children {
				sum += sub.Children[i].Position.Y
			}
			sub.Position.Y = sum / (len(sub.Children))
		}
	}
	sub.Ideal = sub.Position
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
