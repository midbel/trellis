package trellis

import "fmt"

type Dimension struct {
	Width  int
	Height int
}

func NewDimension(w int, h int) Dimension {
	d := Dimension{
		Width:  w,
		Height: h,
	}
	return d
}

func (d *Dimension) Validate() error {
	if d.Width <= 0 {
		return fmt.Errorf("width can not be equal to 0 or negative")
	}
	if d.Height <= 0 {
		return fmt.Errorf("height can not be equal to 0 or negative")
	}
	return nil
}

func (d *Dimension) Valid(x, y int) bool {
	return x >= 0 && x < d.Width && y >= 0 && y < d.Height
}

type Point struct {
	X, Y int
}

func NewPoint(x, y int) Point {
	return Point{
		X: x,
		Y: y,
	}
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

func (p Point) String() string {
	return fmt.Sprintf("point(%d, %d)", p.X, p.Y)
}

type Insets struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

func NewInsets(top, right, bottom, left int) Insets {
	return Insets{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}

func (i Insets) Zero() bool {
	return i.Top == 0 && i.Right == 0 && i.Bottom == 0 && i.Left == 0
}

func (i Insets) H() int {
	return i.Left + i.Right
}

func (i Insets) V() int {
	return i.Top + i.Bottom
}

type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (r Rect) UpperLeft() Point {
	return NewPoint(r.X, r.Y)
}

func (r Rect) MidLeft() Point {
	return NewPoint(r.X, r.Y+r.OffsetY())
}

func (r Rect) MidRight() Point {
	return NewPoint(r.EndX(), r.Y+r.OffsetY())
}

func (r Rect) UpperMid() Point {
	return NewPoint(r.X+r.OffsetX(), r.Y)
}

func (r Rect) LowerMid() Point {
	return NewPoint(r.X+r.OffsetX(), r.EndY())
}

func (r Rect) UpperRight() Point {
	return NewPoint(r.EndX(), r.Y)
}

func (r Rect) LowerRight() Point {
	return NewPoint(r.EndX(), r.EndY())
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

func (r Rect) applyMargins(margin Insets) Rect {
	if margin.Zero() {
		return r
	}
	if r.Width > margin.Left+margin.Right {
		r.Width -= margin.Left + margin.Right
	}
	if r.Height > margin.Top+margin.Bottom {
		r.Height -= margin.Top + margin.Bottom
	}
	r.X += margin.Left
	r.Y += margin.Top
	return r
}
