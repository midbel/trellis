package trellis

type PathType uint8

const (
	ManathanPath PathType = iota
	DirectPath
	CurvePath
)

func ParsePath(str string) (PathType, error) {
	switch str {
	case "direct":
		return DirectPath, nil
	case "manathan", "classic":
		return ManathanPath, nil
	case "curve", "bezier", "cubic":
		return CurvePath, nil
	default:
		return 0, unknown("path", str)
	}
}

type Path struct {
	Start Point
	End   Point
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
		dist  = p.DistanceX()
		mid   = dist / 2
		start Path
		end   Path
		via   Path
		from  Point
		to    Point
	)
	if mid%2 != 0 {
		mid--
	}
	mid = min(p.Start.X+mid, p.End.X-mid)

	from = NewPoint(p.Start.X, p.Start.Y)
	to = NewPoint(mid, p.Start.Y)
	start = NewPath(from, to)

	from = NewPoint(mid, p.End.Y)
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
		dist  = p.DistanceY()
		mid   = dist / 2
		start Path
		end   Path
		via   Path
		from  Point
		to    Point
	)
	if mid%2 != 0 {
		mid--
	}
	mid = min(p.Start.Y+mid, p.End.Y-mid)

	from = NewPoint(p.Start.X, p.Start.Y)
	to = NewPoint(p.Start.X, mid)
	start = NewPath(from, to)

	from = NewPoint(p.End.X, mid)
	to = NewPoint(p.End.X, p.End.Y)
	end = NewPath(from, to)

	if start.Start.X > end.Start.X {
		via = NewPath(end.Start, start.End)
	} else {
		via = NewPath(start.End, end.Start)
	}
	return []Path{start, via, end}
}

func horizontalPath(from, to *Item, opts RenderOptions) Path {
	if !from.Position.BeforeX(to.Position) {
		from, to = to, from
	}
	var (
		start  = from.Position
		end    = to.Position
		offset = from.DisplayWidth()
	)
	start.X += offset + opts.Margin
	end.X -= opts.Margin
	return NewPath(start, end)
}

func verticalPath(from, to *Item, opts RenderOptions) Path {
	if !from.Position.BeforeY(to.Position) {
		from, to = to, from
	}
	var (
		start = from.Position
		end   = to.Position
	)
	start.X += from.Size() / 2
	end.X += to.Size() / 2

	start.Y++
	end.Y--

	return NewPath(start, end)
}
