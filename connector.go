package trellis

import (
	"fmt"
	"slices"
	"strings"
)

type PathType uint8

const (
	DirectPath PathType = 1 << iota
	ManathanPath
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
	Kind PathType

	Controls []Point
	Segments []Segment
}

type Connector struct {
	Start Point
	End   Point
	Path  Path
	
	Paths []Segment
}

func NewConnector(paths []Segment) Connector {
	// if len(paths) == 0 {
	// 	return
	// }
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

func horizontalPath(from, to *Item, opts RenderOptions) Connector {
	if !from.Position.BeforeX(to.Position) {
		from, to = to, from
	}
	var (
		start  = from.Position
		end    = to.Position
		offset = from.DisplayWidth()
	)
	if start.Y == end.Y {
		s := Segment{
			Start: start,
			End:   end,
		}
		s.Start.X += offset
		s.End.X--
		return NewConnector([]Segment{s})
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

	return NewConnector([]Segment{f, v, t})
}

func verticalPath(from, to *Item, opts RenderOptions) Connector {
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
		return NewConnector([]Segment{s})
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
	return NewConnector([]Segment{f, v, t})
}
