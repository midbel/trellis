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

func (d *Dimension) Resize(w, h int) error {
	d.Width = w
	d.Height = h
	return d.Validate()
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
