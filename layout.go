package trellis

import (
	"slices"
)

type ItemsSet struct {
	Width  int
	Height int
	Items  []*Item
}

func (s ItemsSet) Dimension() Dimension {
	return NewDimension(s.Width, s.Height)
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

func (i *Item) Weight() int {
	if i.Leaf() {
		return 1
	}
	var depth int
	for _, c := range i.Children {
		depth += c.Weight()
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
		i.Position.X = i.Bounds.EndX() - i.DisplayWidth()
	default:
		i.Position.X = i.Bounds.StartX() + i.Bounds.OffsetX() - (i.DisplayWidth() / 2)
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
	return i.DisplayWidth()
}

func stdVerticalLayout(root *Node, opts RenderOptions) *ItemsSet {
	var (
		mk  = defaultTreeLayout()
		is  = mk.Make(root, opts)
		set ItemsSet
	)
	for i := range is {
		is[i].Position = is[i].Position.Swap()
		is[i].Ideal = is[i].Position
	}
	var (
		extent = maxFromItems(is, func(i *Item) int { return i.Position.X + opts.Spacing })
		level  = maxFromItems(is, func(i *Item) int { return i.Position.Y })
	)
	ix := slices.IndexFunc(is, func(it *Item) bool {
		return it.Root()
	})
	if ix < 0 {
		return nil
	}
	if extent == 0 {
		extent = opts.Spacing
	}

	// if extent > opts.Size.Width {
	// 	opts.Size.Width = extent
	// }

	computeVerticalCoordinates(is[ix], opts, extent, level)
	set.Width = maxFromItems(is, func(i *Item) int { return i.Bounds.EndX() })
	set.Height = maxFromItems(is, func(i *Item) int { return i.Bounds.EndY() })
	set.Items = is
	return &set
}

func computeVerticalCoordinates(node *Item, opts RenderOptions, spacing, level int) {
	height := opts.Size.Height / (level + 1)

	computeVerticalChildren(node, opts, spacing, level, height)
	resolveVerticalChildren(node, opts)
	computeVerticalNode(node, opts, spacing, height)
}

func computeVerticalChildren(node *Item, opts RenderOptions, spacing, level, height int) {
	for _, x := range node.Children {
		if !x.Leaf() {
			computeVerticalCoordinates(x, opts, spacing, level)
			continue
		}
		var (
			startY = (x.Ideal.Y * height)
			startX = (x.Ideal.X * opts.Size.Width / spacing)
			endX   = ((x.Ideal.X + opts.Spacing) * opts.Size.Width) / spacing
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

func resolveVerticalChildren(node *Item, opts RenderOptions) {
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

func computeVerticalNode(node *Item, opts RenderOptions, spacing, height int) {
	var (
		first = node.FirstLeaf()
		last  = node.LastLeaf()
	)
	if spacing == 0 {
		spacing++
	}
	node.Position.X = node.Ideal.X * opts.Size.Width / spacing
	node.Position.Y = node.Ideal.Y * height

	if node.Leaf() {
		node.Bounds = Rect{
			X:      node.Position.X,
			Y:      node.Position.Y,
			Width:  opts.Size.Width / spacing,
			Height: height,
		}
	} else {
		node.Bounds = Rect{
			X:      first.Bounds.StartX(),
			Y:      node.Position.Y,
			Width:  last.Bounds.EndX() - first.Bounds.StartX(),
			Height: height,
		}
		if len(node.Children) == 1 {
			node.Bounds.Width = opts.estimateMinSize
		}
	}

	node.Bounds = applyMargins(node.Bounds, opts.Margin)

	node.AlignX(opts.AlignX)
	node.AlignY(opts.AlignY)
}

func countLeaves(root *Node) int {
	if root.Leaf() {
		return 1
	}
	var sum int
	for _, n := range root.Nodes {
		sum += countLeaves(n)
	}
	return sum
}

func stdHorizontalLayout(root *Node, opts RenderOptions) *ItemsSet {
	var (
		mk     = defaultTreeLayout()
		is     = mk.Make(root, opts)
		extent = maxFromItems(is, func(i *Item) int { return i.Position.Y + opts.Spacing })
		level  = maxFromItems(is, func(i *Item) int { return i.Position.X })
		set    ItemsSet
	)
	ix := slices.IndexFunc(is, func(it *Item) bool {
		return it.Root()
	})
	if ix < 0 {
		return nil
	}
	if extent == 0 {
		extent = opts.Spacing
	}

	// if extent > opts.Size.Height {
	// 	opts.Size.Height = extent
	// }

	computeHorizontalCoordinates(is[ix], opts, extent, level)
	set.Width = maxFromItems(is, func(i *Item) int { return i.Bounds.EndX() })
	set.Height = maxFromItems(is, func(i *Item) int { return i.Bounds.EndY() })
	set.Height = max(set.Height, opts.Size.Height)
	set.Items = is
	return &set
}

func computeHorizontalCoordinates(node *Item, opts RenderOptions, spacing, level int) {
	width := opts.Size.Width / (level + 1)

	computeHorizontalChildren(node, opts, spacing, level, width)
	resolveHorizontalChildren(node, opts)
	computeHorizontalNode(node, opts, spacing, width)
}

func computeHorizontalChildren(node *Item, opts RenderOptions, spacing, level, width int) {
	for _, x := range node.Children {
		if !x.Leaf() {
			computeHorizontalCoordinates(x, opts, spacing, level)
			continue
		}
		var (
			startX = x.Ideal.X * width
			startY = (x.Ideal.Y * opts.Size.Height / spacing)
			endY   = ((x.Ideal.Y + opts.Spacing) * opts.Size.Height) / spacing
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

func resolveHorizontalChildren(node *Item, opts RenderOptions) {
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

func computeHorizontalNode(node *Item, opts RenderOptions, spacing, width int) {
	var (
		first = node.FirstLeaf()
		last  = node.LastLeaf()
	)
	if spacing == 0 {
		spacing++
	}
	node.Position.X = node.Ideal.X * width
	node.Position.Y = node.Ideal.Y * opts.Size.Height / spacing

	node.Bounds = Rect{
		X:      node.Position.X,
		Y:      first.Bounds.StartY(),
		Width:  width,
		Height: last.Bounds.EndY() - first.Bounds.StartY(),
	}
	if node.Len() == 1 {
		node.Bounds.Height = opts.estimateMinSize // opts.Size.Height
	}

	node.Bounds = applyMargins(node.Bounds, opts.Margin)

	node.AlignX(opts.AlignX)
	node.AlignY(opts.AlignY)
}

const compactBarWidth = 2

func compactLayout(root *Node, opts RenderOptions) *ItemsSet {
	clone := opts.Clone()
	clone.Spacing = 1

	var (
		mk    = defaultTreeLayout()
		items = mk.Make(root, clone)
		set   ItemsSet
	)
	items[0].Bounds = Rect{
		Width:  opts.Size.Width,
		Height: opts.Size.Height,
		X:      items[0].Position.X,
		Y:      items[0].Position.Y,
	}

	for i := 1; i < len(items); i++ {
		for items[i].Position.Y <= items[i-1].Position.Y {
			items[i].Position.Y++
		}
		set.Height = items[i].Position.Y
		items[i].Bounds = Rect{
			X:      items[i].Position.X,
			Y:      items[i].Position.Y,
			Width:  opts.Size.Width - items[i].Position.X,
			Height: 1,
		}
	}
	set.Height++

	ix := slices.IndexFunc(items, func(i *Item) bool {
		return i.Root()
	})
	rearrangeCompactChildren(items[ix], opts.Spacing)

	set.Items = items
	return &set
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

func (m *treeLayout) Make(node *Node, opts RenderOptions) []*Item {
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

func (m *treeLayout) makeLayout(node *Node, depth int, opts RenderOptions) *Item {
	sub := Item{
		Content: opts.Render(node, opts),
		root:    depth == 0,
	}
	sub.Position.X = depth
	depth++
	if opts.MaxDepth == 0 || depth <= opts.MaxDepth {
		for _, n := range node.Nodes {
			child := m.makeLayout(n, depth, opts)
			sub.Children = append(sub.Children, child)
		}
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
			if len(sub.Children) > 0 {
				var sum int
				for i := range sub.Children {
					sum += sub.Children[i].Position.Y
				}
				sub.Position.Y = sum / (len(sub.Children))
			} else {
				sub.Position.Y = m.siblingsSpacing
				m.siblingsSpacing += opts.Spacing
			}
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
