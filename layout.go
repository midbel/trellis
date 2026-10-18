package trellis

import (
	"fmt"
	"slices"
)

type ItemsSet struct {
	Width  int
	Height int
	Items  []*Item
}

func createSet(is []*Item) *ItemsSet {
	return &ItemsSet{
		Items: is,
	}
}

func (s *ItemsSet) Root() (*Item, error) {
	ix := slices.IndexFunc(s.Items, func(i *Item) bool { return i.root })
	if ix < 0 {
		return nil, fmt.Errorf("root not found")
	}
	return s.Items[ix], nil
}

func (s *ItemsSet) Update() {
	s.Width = maxFromItems(s.Items, func(i *Item) int {
		return i.Bounds.EndX()
	})
	s.Height = maxFromItems(s.Items, func(i *Item) int {
		return i.Bounds.EndY()
	})
}

func (s *ItemsSet) Dimension() Dimension {
	return NewDimension(s.Width, s.Height)
}

type Item struct {
	Content

	Metric Metrics

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
	var count int
	for _, c := range i.Children {
		count += c.Weight()
	}
	return count + 1
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

func (i *Item) X() int {
	var (
		bounds = i.InnerBounds()
		left   int
	)
	if i.Padding.Left == 0 {
		left++
	}
	return bounds.X + i.Padding.Left + left
}

func (i *Item) Y() int {
	bounds := i.InnerBounds()
	return bounds.Y + i.Padding.Top
}

func (i *Item) OuterBounds() Rect {
	r := i.Bounds
	if i.Margin.V() > r.Height || i.Margin.H() > r.Width {
		return r
	}
	r.X += i.Margin.Left
	r.Y += i.Margin.Top
	r.Width -= i.Margin.H()
	r.Height -= i.Margin.V()
	return r
}

func (i *Item) InnerBounds() Rect {
	var (
		bounds = i.OuterBounds()
		left   int
	)
	if i.Padding.H() == 0 {
		left++
	}
	r := Rect{
		X:      bounds.X,
		Y:      bounds.Y,
		Width:  i.Metric.Width(i.Value) + i.Padding.H() + left,
		Height: i.Metric.Height() + i.Padding.V(),
	}
	switch i.AlignX {
	case AlignStart:
	case AlignEnd:
		r.X = bounds.EndX() - r.Width
	default:
		r.X += (bounds.Width - r.Width) / 2
	}
	switch i.AlignY {
	case AlignStart:
	case AlignEnd:
		r.Y = bounds.EndY() - r.Height
	default:
		r.Y += (bounds.Height - r.Height) / 2
	}
	return r
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
	return i.Metric.Width(i.Value)
}

func stdVerticalLayout(root *Item, opts RenderOptions) (*ItemsSet, error) {
	var (
		mk  = defaultTreeLayout()
		is  = mk.Make(root, opts)
		set = createSet(is)
	)
	for i := range set.Items {
		set.Items[i].Position = set.Items[i].Position.Swap()
		set.Items[i].Ideal = set.Items[i].Position
	}
	var (
		extent = maxFromItems(set.Items, func(i *Item) int { return i.Position.X + opts.Spacing })
		level  = maxFromItems(set.Items, func(i *Item) int { return i.Position.Y })
	)
	if extent == 0 {
		extent = opts.Spacing
	}
	root, err := set.Root()
	if err != nil {
		return nil, err
	}

	computeVerticalCoordinates(root, opts, extent, level)
	set.Update()
	return set, nil
}

func computeVerticalCoordinates(node *Item, opts RenderOptions, extent, level int) {
	height := opts.Size.Height / (level + 1)

	computeVerticalChildren(node, opts, extent, level, height)
	resolveVerticalChildren(node)
	computeVerticalNode(node, opts, extent, height)
}

func computeVerticalChildren(node *Item, opts RenderOptions, extent, level, height int) {
	for _, x := range node.Children {
		if !x.Leaf() {
			computeVerticalCoordinates(x, opts, extent, level)
			continue
		}
		var (
			startY = (x.Ideal.Y * height)
			startX = (x.Ideal.X * opts.Size.Width / extent)
			endX   = ((x.Ideal.X + opts.Spacing) * opts.Size.Width) / extent
		)
		x.Position.X = startX
		x.Position.Y = startY
		x.Bounds = Rect{
			X:      startX,
			Y:      startY,
			Width:  endX - startX,
			Height: height,
		}

		if x.Bounds.Width < opts.Spacing+1 {
			x.Bounds.Width += opts.Spacing + 1
		}
	}
}

func resolveVerticalChildren(node *Item) {
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

func computeVerticalNode(node *Item, opts RenderOptions, extent, height int) {
	var (
		first = node.FirstLeaf()
		last  = node.LastLeaf()
	)
	if extent == 0 {
		extent++
	}
	node.Position.X = node.Ideal.X * opts.Size.Width / extent
	node.Position.Y = node.Ideal.Y * height

	if node.Leaf() {
		node.Bounds = Rect{
			X:      node.Position.X,
			Y:      node.Position.Y,
			Width:  opts.Size.Width / extent,
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
}

func stdHorizontalLayout(root *Item, opts RenderOptions) (*ItemsSet, error) {
	var (
		mk     = defaultTreeLayout()
		is     = mk.Make(root, opts)
		extent = maxFromItems(is, func(i *Item) int { return i.Position.Y + opts.Spacing })
		level  = maxFromItems(is, func(i *Item) int { return i.Position.X })
		set    = createSet(is)
	)
	if extent == 0 {
		extent = opts.Spacing
	}
	root, err := set.Root()
	if err != nil {
		return nil, err
	}

	computeHorizontalCoordinates(root, opts, extent, level)
	set.Update()
	set.Height = max(set.Height, opts.Size.Height)
	return set, nil
}

func computeHorizontalCoordinates(node *Item, opts RenderOptions, extent, level int) {
	width := opts.Size.Width / (level + 1)

	computeHorizontalChildren(node, opts, extent, level, width)
	resolveHorizontalChildren(node)
	computeHorizontalNode(node, opts, extent, width)
}

func computeHorizontalChildren(node *Item, opts RenderOptions, extent, level, width int) {
	for _, x := range node.Children {
		if !x.Leaf() {
			computeHorizontalCoordinates(x, opts, extent, level)
			continue
		}
		var (
			startX = x.Ideal.X * width
			startY = (x.Ideal.Y * opts.Size.Height / extent)
			endY   = ((x.Ideal.Y + opts.Spacing) * opts.Size.Height) / extent
		)
		x.Position.X = startX
		x.Position.Y = startY
		x.Bounds = Rect{
			X:      startX,
			Y:      startY,
			Width:  width,
			Height: endY - startY,
		}

		if x.Bounds.Height < opts.Spacing+1 {
			x.Bounds.Height += opts.Spacing + 1
		}
	}
}

func resolveHorizontalChildren(node *Item) {
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

func computeHorizontalNode(node *Item, opts RenderOptions, extent, width int) {
	var (
		first = node.FirstLeaf()
		last  = node.LastLeaf()
	)
	if extent == 0 {
		extent++
	}
	node.Position.X = node.Ideal.X * width
	node.Position.Y = node.Ideal.Y * opts.Size.Height / extent

	node.Bounds = Rect{
		X:      node.Position.X,
		Y:      first.Bounds.StartY(),
		Width:  width,
		Height: last.Bounds.EndY() - first.Bounds.StartY(),
	}
	if node.Len() == 1 {
		node.Bounds.Height = opts.estimateMinSize
	}
}

func compactLayout(root *Item, opts RenderOptions) (*ItemsSet, error) {
	clone := opts.Clone()
	clone.Spacing = 1

	var (
		mk    = defaultTreeLayout()
		items = mk.Make(root, clone)
		set   = createSet(items)
	)
	items[0].Bounds = Rect{
		Width:  opts.Size.Width,
		Height: opts.Size.Height,
		X:      items[0].Position.X,
		Y:      items[0].Position.Y,
	}

	for i := 1; i < len(set.Items); i++ {
		for set.Items[i].Position.Y <= set.Items[i-1].Position.Y {
			set.Items[i].Position.Y++
		}
		set.Height = set.Items[i].Position.Y
		set.Items[i].Bounds = Rect{
			X:      set.Items[i].Position.X,
			Y:      set.Items[i].Position.Y,
			Width:  opts.Size.Width - set.Items[i].Position.X,
			Height: 1,
		}
	}
	set.Width = opts.Size.Width
	set.Height++

	root, err := set.Root()
	if err != nil {
		return nil, err
	}
	rearrangeCompactChildren(root, opts.Spacing)
	return set, nil
}

func rearrangeCompactChildren(node *Item, extent int) {
	for i := range node.Children {
		node.Children[i].Position.X = node.Position.X + extent
		rearrangeCompactChildren(node.Children[i], extent)
	}
}

type treeLayout struct {
	siblingsSpacing int
	levelSpacing    int
}

func defaultTreeLayout() *treeLayout {
	return &treeLayout{}
}

func (m *treeLayout) Make(node *Item, opts RenderOptions) []*Item {
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

func (m *treeLayout) makeLayout(node *Item, depth int, opts RenderOptions) *Item {
	node.Position.X = depth
	depth++
	if opts.MaxDepth == 0 || depth <= opts.MaxDepth {
		children := node.Children
		node.Children = node.Children[:0]
		for _, n := range children {
			child := m.makeLayout(n, depth, opts)
			node.Children = append(node.Children, child)
		}
	}
	if node.Leaf() {
		node.Position.Y = m.siblingsSpacing
		m.siblingsSpacing += opts.Spacing
	} else {
		if opts.Placement == AlignStart {
			node.Position.Y = node.Children[0].Position.Y
		} else if opts.Placement == AlignEnd {
			node.Position.Y = node.Children[len(node.Children)-1].Position.Y
		} else {
			if len(node.Children) > 0 {
				var sum int
				for i := range node.Children {
					sum += node.Children[i].Position.Y
				}
				node.Position.Y = sum / (len(node.Children))
			} else {
				node.Position.Y = m.siblingsSpacing
				m.siblingsSpacing += opts.Spacing
			}
		}
	}
	node.Ideal = node.Position
	m.levelSpacing = max(depth-1, m.levelSpacing)
	return node
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
