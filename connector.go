package trellis

import (
	"fmt"
	"slices"
	"strings"
)

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
	case "manathan":
		return ManathanPath, nil
	case "curve":
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

func (p Path) X() int {
	return p.Start.X
}

func (p Path) Y() int {
	return p.Start.Y
}

func (p Path) Move(x, y int) Path {
	p.Start.X += x
	p.Start.Y += y
	p.End.X += x
	p.End.Y += y
	return p
}

type Connector struct {
	Start Point
	End   Point
	Path  Path

	Paths []Segment
}

func NewConnector(paths []Segment) Connector {
	return Connector{
		Paths: paths,
	}
}

func (c Connector) Move(x, y int) Connector {
	cp := NewConnector(slices.Clone(c.Paths))
	for i := range cp.Paths {
		cp.Paths[i].Start.X += x
		cp.Paths[i].Start.Y += y
		cp.Paths[i].End.X += x
		cp.Paths[i].End.Y += y
	}
	return cp
}

func (c Connector) X() int {
	return c.Paths[0].Start.X
}

func (c Connector) Y() int {
	return c.Paths[0].Start.Y
}

func (c Connector) String() string {
	var str strings.Builder
	str.WriteString("connector(")
	for i, p := range c.Paths {
		if i > 0 {
			str.WriteRune(',')
			str.WriteRune(' ')
		}
		str.WriteString(p.String())
	}
	str.WriteString(")")
	return str.String()
}

type Segment struct {
	Start Point
	End   Point
}

func NewSegment(start, end Point) Segment {
	return Segment{
		Start: start,
		End:   end,
	}
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

func (s Segment) String() string {
	return fmt.Sprintf("segment(%s -> %s)", s.Start, s.End)
}

func horizontalPath(from, to *Item) Path {
	if !from.Position.BeforeX(to.Position) {
		from, to = to, from
	}
	var (
		start  = from.Position
		end    = to.Position
		offset = from.DisplayWidth()
	)
	start.X += offset
	end.X--
	return NewPath(start, end)
}

func verticalPath(from, to *Item) Path {
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
