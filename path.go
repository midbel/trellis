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

type Path struct {
	Start Point
	End   Point
	Pivot int
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
	to = NewPoint(p.Pivot, p.Start.Y)
	start = NewPath(from, to)

	from = NewPoint(p.Pivot, p.End.Y)
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
	to = NewPoint(p.Start.X, p.Pivot)
	start = NewPath(from, to)

	from = NewPoint(p.End.X, p.Pivot)
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

	if to.Bounds.X == from.Bounds.EndX() {
		p.Pivot = to.Bounds.X
	} else {
		diff := (to.Bounds.X - from.Bounds.EndX()) / 2
		p.Pivot = from.Bounds.EndX() + diff
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

	if to.Bounds.Y == from.Bounds.EndY() {
		p.Pivot = to.Bounds.Y
	} else {
		diff := (to.Bounds.Y - from.Bounds.EndY()) / 2
		p.Pivot = from.Bounds.EndY() + diff
	}
	return p
}
