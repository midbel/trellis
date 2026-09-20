package trellis

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"github.com/midbel/angle/xml"
)

const space = ' '

type View interface {
	Render(io.Writer) error
	put(int, int, Cell) error
}

type XmlFile struct {
	root *xml.Element
}

func NewXml(opts *Options) (View, error) {
	el := xml.Element{
		Name: xml.NewName("tree"),
		Attributes: []xml.Attribute{
			{
				Name:  xml.NewName("width"),
				Value: strconv.Itoa(opts.Width),
			},
			{
				Name:  xml.NewName("height"),
				Value: strconv.Itoa(opts.Height),
			},
			{
				Name:  xml.NewName("orientation"),
				Value: opts.Orient.String(),
			},
		},
	}
	return &XmlFile{
		root: &el,
	}, nil
}

func (f *XmlFile) Render(w io.Writer) error {
	var (
		doc = xml.NewDocument(*f.root)
		enc = xml.NewEncoder(w)
	)
	return enc.Encode(doc)
}

func (f *XmlFile) put(x, y int, cell Cell) error {
	switch c := cell.(type) {
	case Content:
		f.createElementForContent(x, y, c)
	case Connector:
		f.createElementForConnector(x, y, c)
	default:
	}
	return nil
}

func (f *XmlFile) createElementForContent(x, y int, val Content) {
	el := xml.Element{
		Name: xml.NewName("content"),
		Attributes: []xml.Attribute{
			{
				Name:  xml.NewName("x"),
				Value: strconv.Itoa(x),
			},
			{
				Name:  xml.NewName("y"),
				Value: strconv.Itoa(y),
			},
		},
		Children: []xml.Node{
			xml.Text{Value: string(val.Value)},
		},
	}
	f.root.Children = append(f.root.Children, el)
}

func (f *XmlFile) createElementForConnector(x, y int, conn Connector) {
	el := xml.Element{
		Name: xml.NewName("connector"),
	}
	for _, seg := range conn.Paths {
		sub := f.createElementForSegment(seg)
		el.Children = append(el.Children, sub)
	}
	f.root.Children = append(f.root.Children, el)
}

func (f *XmlFile) createElementForSegment(seg Segment) xml.Element {
	return xml.Element{
		Name: xml.NewName("segment"),
		Children: []xml.Node{
			xml.Element{
				Name: xml.NewName("start"),
				Attributes: []xml.Attribute{
					{
						Name:  xml.NewName("x"),
						Value: strconv.Itoa(seg.Start.X),
					},
					{
						Name:  xml.NewName("y"),
						Value: strconv.Itoa(seg.Start.Y),
					},
				},
			},
			xml.Element{
				Name: xml.NewName("end"),
				Attributes: []xml.Attribute{
					{
						Name:  xml.NewName("x"),
						Value: strconv.Itoa(seg.End.X),
					},
					{
						Name:  xml.NewName("y"),
						Value: strconv.Itoa(seg.End.Y),
					},
				},
			},
		},
	}
}

type Screen struct {
	lines [][]rune
	dim   Dimension

	connector ConnectorStyle
	border    bool

	crossings []Point
}

func NewScreen(opts *Options) (View, error) {
	sc := &Screen{
		lines: make([][]rune, opts.Height),
		dim: Dimension{
			Width:  opts.Width,
			Height: opts.Height,
		},
		border:    opts.Border,
		connector: opts.Style,
	}
	if err := sc.dim.Validate(); err != nil {
		return nil, err
	}
	for i := range sc.lines {
		sc.lines[i] = make([]rune, opts.Width)
	}
	sc.fillGrid()
	return sc, nil
}

func (s *Screen) Render(w io.Writer) error {
	ws := bufio.NewWriter(w)
	if s.border {
		if _, err := ws.WriteRune(s.connector.TopLeft()); err != nil {
			return err
		}
		for range s.dim.Width {
			if _, err := ws.WriteRune(s.connector.HorizontalBar()); err != nil {
				return err
			}
		}
		if _, err := ws.WriteRune(s.connector.TopRight()); err != nil {
			return err
		}
		if _, err := ws.WriteRune('\n'); err != nil {
			return err
		}
	}
	s.writeCrossings()
	for i := range s.lines {
		if s.border {
			if _, err := ws.WriteRune(s.connector.VerticalBar()); err != nil {
				return err
			}
		}
		for j := range s.lines[i] {
			_, err := ws.WriteRune(s.lines[i][j])
			if err != nil {
				return err
			}
		}
		if s.border {
			if _, err := ws.WriteRune(s.connector.VerticalBar()); err != nil {
				return err
			}
		}
		if _, err := ws.WriteRune('\n'); err != nil {
			return err
		}
	}
	if s.border {
		if _, err := ws.WriteRune(s.connector.BottomLeft()); err != nil {
			return err
		}
		for range s.dim.Width {
			if _, err := ws.WriteRune(s.connector.HorizontalBar()); err != nil {
				return err
			}
		}
		if _, err := ws.WriteRune(s.connector.BottomRight()); err != nil {
			return err
		}
		if _, err := ws.WriteRune('\n'); err != nil {
			return err
		}
	}
	return ws.Flush()
}

func (s *Screen) put(x, y int, cell Cell) error {
	var err error
	if !s.dim.Valid(x, y) {
		return fmt.Errorf("invalid coordinates (%d, %d)", x, y)

	}
	switch c := cell.(type) {
	case Content:
		err = s.putContent(x, y, c)
	case Connector:
		err = s.putConnector(x, y, c)
	default:
	}
	return err
}

func (s *Screen) writeCrossings() {
	isBar := func(y, x int) bool {
		if !s.dim.Valid(x, y) {
			return false
		}
		return s.lines[y][x] != space
	}
	for _, p := range s.crossings {
		if !s.dim.Valid(p.X, p.Y) {
			continue
		}
		var (
			left   = isBar(p.Y, p.X-1)
			right  = isBar(p.Y, p.X+1)
			top    = isBar(p.Y-1, p.X)
			bottom = isBar(p.Y+1, p.X)
			char   = s.lines[p.Y][p.X]
		)
		switch {
		case left && right && top && bottom:
			// all four
			char = s.connector.CrossPath()
		case left && right && top && !bottom:
			// horizontal up
			char = s.connector.HorizontalUp()
		case left && right && bottom && !top:
			// horizontal down
			char = s.connector.HorizontalDown()
		case top && bottom && left && !right:
			// vertical left
			char = s.connector.VerticalLeft()
		case top && bottom && right && !left:
			// vertical right
			char = s.connector.VerticalRight()
		case bottom && left && !right && !top:
			// bottom left
			char = s.connector.TopRight()
		case top && left && !right && !bottom:
			// top left
			char = s.connector.BottomRight()
		case bottom && right && !left && !top:
			// bottom right
			char = s.connector.TopLeft()
		case top && right && !left && !bottom:
			// top right
			char = s.connector.BottomLeft()
		}
		if s.dim.Valid(p.X, p.Y) {
			s.lines[p.Y][p.X] = char
		}
	}
}

func (s *Screen) putContent(x, y int, val Content) error {
	for _, r := range val.Value {
		s.writeChar(x, y, r)
		x += RuneWidth(r)
	}
	return nil
}

func (s *Screen) putConnector(x, y int, conn Connector) error {
	for _, seg := range conn.Paths {
		s.crossings = append(s.crossings, seg.Start, seg.End)
		if seg.Horizontal() {
			s.horizontalConnector(seg)
		} else {
			s.verticalConnector(seg)
		}
	}
	return nil
}

func (s *Screen) verticalConnector(seg Segment) {
	char := s.connector.VerticalBar()
	if seg.DistanceY() == 1 {
		s.writeSymbol(seg.Start.X, seg.Start.Y, char)
		return
	}
	var (
		start = seg.Start
		end   = seg.End
	)
	if start.BeforeY(end) {
		start, end = end, start
	}
	for y := start.Y; y >= end.Y; y-- {
		s.writeSymbol(start.X, y, char)
	}
}

func (s *Screen) horizontalConnector(seg Segment) {
	char := s.connector.HorizontalBar()
	if seg.DistanceX() == 1 {
		s.writeSymbol(seg.Start.X, seg.Start.Y, char)
		return
	}
	var (
		start = seg.Start
		end   = seg.End
	)
	if end.BeforeX(start) {
		start, end = end, start
	}
	for x := start.X; x <= end.X; x++ {
		s.writeSymbol(x, start.Y, char)
	}
}

func (s *Screen) writeChar(x, y int, char rune) {
	if !s.dim.Valid(x, y) {
		return
	}
	s.lines[y][x] = char
}

func (s *Screen) writeSymbol(x, y int, char rune) {
	s.writeChar(x, y, char)
}

func (s *Screen) fillGrid() {
	for i := range s.lines {
		for j := range s.lines[i] {
			s.lines[i][j] = space
		}
	}
}
