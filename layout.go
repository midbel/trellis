package trellis

import (
	"fmt"
	"slices"
	"strings"
)

type LayoutFunc func(*Node, *Options) []*Item

func Layout(orient Orientation) (LayoutFunc, error) {
	var fn LayoutFunc
	switch orient {
	case HorizontalLayout:
		fn = stdHorizontalLayout
	case VerticalLayout:
		fn = stdVerticalLayout
	case CompactLayout:
		fn = compactLayout
	default:
		return nil, fmt.Errorf("unsupported layout")
	}
	return fn, nil
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
	Bounds   Rect
}

func ComputeLayout(root *Node, options *Options) (CoordinateMap, error) {
	opts, err := prepareOptions(options)
	if err != nil {
		return CoordinateMap{}, err
	}
	var (
		is  []*Item
		res CoordinateMap
		fn  LayoutFunc
	)
	fn, err = Layout(options.Orient)
	if err != nil {
		return res, err
	}
	is = fn(root, options)
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
			Bounds: is[i].Bounds,
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
	f.End.X = from.Bounds.EndX() + opts.Margin

	t := Segment{
		Start: end,
		End:   end,
	}
	t.Start.X = to.Bounds.StartX() - opts.Margin
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
	f.End.Y = from.Bounds.EndY() + opts.Margin

	t := Segment{
		Start: end,
		End:   end,
	}
	t.Start.Y = to.Bounds.StartY() - opts.Margin
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

func applyMargins(rect Rect, margin int) Rect {
	if margin == 0 {
		return rect
	}
	if rect.Width > margin+margin {
		rect.Width -= margin + margin
	}
	if rect.Height > margin+margin {
		rect.Height -= margin + margin
	}
	rect.X += margin
	rect.Y += margin
	return rect
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

func (r Rect) OffsetX() int {
	return r.Width / 2
}

func (r Rect) OffsetY() int {
	return r.Height / 2
}

type Item struct {
	Content

	Ideal    Point
	Position Point
	Bounds   Rect

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

func (i *Item) Depth() int {
	if i.Leaf() {
		return 1
	}
	var depth int
	for _, c := range i.Children {
		depth += c.Depth()
	}
	return depth + 1
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
		i.Position.Y = i.Bounds.StartY()
	case AlignEnd:
		i.Position.Y = i.Bounds.EndY()
	default:
		i.Position.Y = i.Bounds.StartY() + i.Bounds.OffsetY()
	}
}

func (i *Item) AlignX(align Alignment) {
	switch align {
	case AlignStart:
		i.Position.X = i.Bounds.StartX()
	case AlignEnd:
		i.Position.X = i.Bounds.EndX() - len(i.Value)
	default:
		i.Position.X = i.Bounds.StartX() + i.Bounds.OffsetX() - (len(i.Value) / 2)
	}
}

func (i *Item) MoveX(delta int) {
	i.Position.X += delta
	i.Bounds.X += delta

	for _, c := range i.Children {
		c.MoveX(delta)
	}
}

func (i *Item) MoveY(delta int) {
	i.Position.Y += delta
	i.Bounds.Y += delta

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

func stdVerticalLayout(root *Node, opts *Options) []*Item {
	var (
		mk = defaultTreeLayout()
		is = mk.Make(root, opts)
	)
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

	opts.Width = maxFromItems(is, func(i *Item) int { return i.Bounds.EndX() })
	opts.Height = maxFromItems(is, func(i *Item) int { return i.Bounds.EndY() })
	return is
}

func computeVerticalCoordinates(node *Item, opts *Options, spacing, level int) {
	height := opts.Height / (level + 1)

	computeVerticalChildren(node, opts, spacing, level, height)
	resolveVerticalChildren(node, opts)
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
		x.Bounds = Rect{
			X:      startX,
			Y:      startY,
			Width:  endX - startX,
			Height: height,
		}
		x.Bounds = applyMargins(x.Bounds, opts.Margin)

		if x.Bounds.Width < opts.Spacing+1 {
			x.Bounds.Width += opts.Spacing + 1
		}
		x.AlignX(opts.AlignX)
		x.AlignY(opts.AlignY)
	}
}

func resolveVerticalChildren(node *Item, opts *Options) {
	if len(node.Children) < 1 {
		return
	}
	boundary := node.Children[0].Bounds.EndX()
	for _, c := range node.Children[1:] {
		if c.Bounds.StartX() < boundary {
			c.MoveX(boundary - c.Bounds.StartX() + 1)
		}
		boundary = c.Bounds.EndX()
	}
}

func computeVerticalNode(node *Item, opts *Options, spacing, height int) {
	var (
		first = node.FirstLeaf()
		last  = node.LastLeaf()
	)
	if spacing == 0 {
		spacing++
	}
	node.Position.X = node.Ideal.X * opts.Width / spacing
	node.Position.Y = node.Ideal.Y * height

	node.Bounds = Rect{
		X:      first.Bounds.StartX(),
		Y:      node.Position.Y,
		Width:  last.Bounds.EndX() - first.Bounds.StartX(),
		Height: height,
	}
	node.Bounds = applyMargins(node.Bounds, opts.Margin)

	node.AlignX(opts.AlignX)
	node.AlignY(opts.AlignY)
}

func stdHorizontalLayout(root *Node, opts *Options) []*Item {
	var (
		mk      = defaultTreeLayout()
		is      = mk.Make(root, opts)
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
	opts.Width = maxFromItems(is, func(i *Item) int { return i.Bounds.EndX() })
	opts.Height = maxFromItems(is, func(i *Item) int { return i.Bounds.EndY() })
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
		x.Position.X = startX
		x.Position.Y = startY
		x.Bounds = Rect{
			X:      startX,
			Y:      startY,
			Width:  width,
			Height: endY - startY,
		}
		x.Bounds = applyMargins(x.Bounds, opts.Margin)

		if x.Bounds.Height < opts.Spacing+1 {
			x.Bounds.Height += opts.Spacing + 1
		}
		x.AlignX(opts.AlignX)
		x.AlignY(opts.AlignY)
	}
}

func resolveHorizontalChildren(node *Item, opts *Options) {
	if len(node.Children) < 1 {
		return
	}
	boundary := node.Children[0].Bounds.EndY()
	for _, c := range node.Children[1:] {
		if c.Bounds.StartY() < boundary {
			c.MoveY(boundary - c.Bounds.StartY() + 1)
		}
		boundary = c.Bounds.EndY()
	}
}

func computeHorizontalNode(node *Item, opts *Options, spacing, width int) {
	var (
		first = node.FirstLeaf()
		last  = node.LastLeaf()
	)
	if spacing == 0 {
		spacing++
	}
	node.Position.X = node.Ideal.X * width
	node.Position.Y = node.Ideal.Y * opts.Height / spacing

	node.Bounds = Rect{
		X:      node.Position.X,
		Y:      first.Bounds.StartY(),
		Width:  width,
		Height: last.Bounds.EndY() - first.Bounds.StartY(),
	}
	node.Bounds = applyMargins(node.Bounds, opts.Margin)

	node.AlignX(opts.AlignX)
	node.AlignY(opts.AlignY)
}

func compactLayout(root *Node, opts *Options) []*Item {
	clone := opts.Clone()
	clone.Spacing = 1

	var (
		mk    = defaultTreeLayout()
		items = mk.Make(root, clone)
	)
	items[0].Bounds = Rect{
		Width:  opts.Width,
		Height: opts.Height,
		X:      items[0].Position.X,
		Y:      items[0].Position.Y,
	}

	for i := 1; i < len(items); i++ {
		for items[i].Position.Y <= items[i-1].Position.Y {
			items[i].Position.Y++
		}
		opts.Height = items[i].Position.Y
		items[i].Bounds = Rect{
			X:      items[i].Position.X,
			Y:      items[i].Position.Y,
			Width:  opts.Width - items[i].Position.X,
			Height: 1,
		}
	}
	opts.Height++

	ix := slices.IndexFunc(items, func(i *Item) bool {
		return i.Root()
	})
	rearrangeCompactChildren(items[ix], opts.Spacing)
	return items
}

func rearrangeCompactChildren(node *Item, spacing int) {
	for i := range node.Children {
		node.Children[i].Position.X = node.Position.X + spacing
		rearrangeCompactChildren(node.Children[i], spacing)
	}
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
	var (
		root = m.makeLayout(node, 0, opts)
		res  = m.flatten(root)
	)
	if opts.Reverse {
		level := m.Depth() - 1
		for i := range res {
			res[i].Ideal.X = level - res[i].Position.X
			res[i].Position = res[i].Ideal
		}
	}
	return res
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
