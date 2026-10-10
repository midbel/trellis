package trellis

type PathType uint8

const (
	ManatthanPath PathType = iota
	DirectPath
	CurvePath
)

func ParsePath(str string) (PathType, error) {
	switch str {
	case "direct":
		return DirectPath, nil
	case "manatthan", "classic":
		return ManatthanPath, nil
	case "curve", "bezier", "cubic":
		return CurvePath, nil
	default:
		return 0, unknown("path", str)
	}
}

type Axis struct {
	Orient Orientation
	Pos    int
}

type Path struct {
	Start Point
	End   Point
	Pivot Axis
}

func NewPath(start, end Point) Path {
	return Path{
		Start: start,
		End:   end,
	}
}

func (p Path) Move(x, y int) Path {
	p.Start.X += x
	p.Start.Y += y
	p.End.X += x
	p.End.Y += y
	if p.Pivot.Orient == HorizontalLayout {
		p.Pivot.Pos += x
	} else {
		p.Pivot.Pos += y
	}
	return p
}

func (p Path) Vertical() bool {
	return p.Start.X == p.End.X
}

func (p Path) Horizontal() bool {
	return p.Start.Y == p.End.Y
}

func (p Path) X() int {
	return p.Start.X
}

func (p Path) Y() int {
	return p.Start.Y
}

func (p Path) DistanceX() int {
	return p.End.X - p.Start.X
}

func (p Path) DistanceY() int {
	return p.End.Y - p.Start.Y
}

func splitPathH(p Path) []Path {
	if p.Start.Y == p.End.Y {
		return []Path{p}
	}
	var (
		start Path
		end   Path
		via   Path
		from  Point
		to    Point
	)
	from = NewPoint(p.Start.X, p.Start.Y)
	to = NewPoint(p.Pivot.Pos, p.Start.Y)
	start = NewPath(from, to)

	from = NewPoint(p.Pivot.Pos, p.End.Y)
	to = NewPoint(p.End.X, p.End.Y)
	end = NewPath(from, to)

	if start.Start.Y > end.Start.Y {
		via = NewPath(end.Start, start.End)
	} else {
		via = NewPath(start.End, end.Start)
	}
	return []Path{start, via, end}
}

func splitPathV(p Path) []Path {
	if p.Start.X == p.End.X {
		return []Path{p}
	}
	var (
		start Path
		end   Path
		via   Path
		from  Point
		to    Point
	)

	from = NewPoint(p.Start.X, p.Start.Y)
	to = NewPoint(p.Start.X, p.Pivot.Pos)
	start = NewPath(from, to)

	from = NewPoint(p.End.X, p.Pivot.Pos)
	to = NewPoint(p.End.X, p.End.Y)
	end = NewPath(from, to)

	if start.Start.X > end.Start.X {
		via = NewPath(end.Start, start.End)
	} else {
		via = NewPath(start.End, end.Start)
	}
	return []Path{start, via, end}
}

func horizontalPath(from, to *Item) Path {
	if from.Bounds.X > to.Bounds.X {
		from, to = to, from
	}
	var (
		start = from.ContentBounds()
		end   = to.ContentBounds()
	)

	ep := end.UpperLeft()
	ep.X--
	p := NewPath(start.UpperRight(), ep)

	p.Pivot.Orient = HorizontalLayout
	if to.Bounds.X == from.Bounds.EndX() {
		p.Pivot.Pos = to.Bounds.X
	} else {
		diff := (to.Bounds.X - from.Bounds.EndX()) / 2
		p.Pivot.Pos = from.Bounds.EndX() + diff
	}
	return p
}

func verticalPath(from, to *Item) Path {
	if from.Bounds.Y > to.Bounds.Y {
		from, to = to, from
	}
	var (
		start = from.ContentBounds()
		end   = to.ContentBounds()
	)
	ep := end.UpperMid()
	ep.Y--
	p := NewPath(start.LowerMid(), ep)

	p.Pivot.Orient = VerticalLayout
	if to.Bounds.Y == from.Bounds.EndY() {
		p.Pivot.Pos = to.Bounds.Y
	} else {
		diff := (to.Bounds.Y - from.Bounds.EndY()) / 2
		p.Pivot.Pos = from.Bounds.EndY() + diff
	}
	return p
}
