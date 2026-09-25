package trellis

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/midbel/angle/svg"
	"github.com/midbel/angle/xml"
	"github.com/midbel/curly"
)

const space = ' '

type View interface {
	Render(io.Writer) error
	put(int, int, Cell) error
}

type JsonFile struct {
	root map[string]any
	opts JsonOptions
}

func NewJson(opts JsonOptions, orient Orientation, size Dimension) (View, error) {
	j := &JsonFile{
		root: make(map[string]any),
	}
	j.root["width"] = size.Width
	j.root["height"] = size.Height
	j.root["orientation"] = orient.String()
	j.root["cells"] = []any{}
	j.root["connectors"] = []any{}
	j.root["canvas"] = []any{}
	return j, nil
}

func (f *JsonFile) Render(w io.Writer) error {
	var ws *curly.Writer
	if f.opts.Compact {
		ws = curly.Compact(w)
	} else {
		ws = curly.NewWriter(w)
	}
	return ws.Write(f.root)
}

func (f *JsonFile) put(x, y int, cell Cell) error {
	switch c := cell.(type) {
	case Content:
		f.appendContent(x, y, c)
	case Connector:
		f.appendConnector(x, y, c)
	case *Canvas:
		f.appendCanvas(x, y, c)
	default:
	}
	return nil
}

func (f *JsonFile) appendContent(x, y int, c Content) {
	v := map[string]any{
		"content": string(c.Value),
		"x":       x,
		"y":       y,
	}
	vs, ok := f.root["cells"].([]any)
	if ok {
		f.root["cells"] = append(vs, v)
	}
}

func (f *JsonFile) appendConnector(x, y int, c Connector) {
	var list []any
	for _, p := range c.Paths {
		s := map[string]any{
			"x": p.Start.X,
			"y": p.Start.Y,
		}
		e := map[string]any{
			"x": p.End.X,
			"y": p.End.Y,
		}
		g := map[string]any{
			"start": s,
			"end":   e,
		}
		list = append(list, g)
	}
	vs, ok := f.root["connectors"].([]any)
	if ok {
		f.root["connectors"] = append(vs, list)
	}
}

func (f *JsonFile) appendCanvas(x, y int, c *Canvas) {
	root := map[string]any{
		"x":          x,
		"y":          y,
		"cells":      []any{},
		"connectors": []any{},
	}
	tmp := f.root
	f.root = root

	for _, p := range c.cells {
		f.put(p.X, p.Y, p.Cell)
	}

	vs, ok := tmp["canvas"].([]any)
	if ok {
		f.root = tmp
		f.root["canvas"] = append(vs, root)
	}
}

type XmlFile struct {
	root *xml.Element
	opts XmlOptions
}

func NewXml(opts XmlOptions, orient Orientation, size Dimension) (View, error) {
	el := xml.Element{
		Name: xml.NewName("canvas"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("width"), strconv.Itoa(size.Width)),
			xml.NewAttribute(xml.NewName("height"), strconv.Itoa(size.Height)),
			xml.NewAttribute(xml.NewName("orientation"), orient.String()),
		},
	}
	return &XmlFile{
		root: &el,
	}, nil
}

func (f *XmlFile) Render(w io.Writer) error {
	var (
		doc = xml.NewDocument(f.root)
		enc = xml.NewEncoder(w)
	)
	enc.SetCompact(f.opts.Compact)
	return enc.Encode(doc)
}

func (f *XmlFile) put(x, y int, cell Cell) error {
	switch c := cell.(type) {
	case Content:
		f.createElementForContent(x, y, c)
	case Connector:
		f.createElementForConnector(x, y, c)
	case *Canvas:
		f.createElementForCanvas(x, y, c)
	default:
	}
	return nil
}

func (f *XmlFile) createElementForCanvas(x, y int, cv *Canvas) {
	el := &xml.Element{
		Name: xml.NewName("canvas"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("x"), strconv.Itoa(x)),
			xml.NewAttribute(xml.NewName("y"), strconv.Itoa(y)),
		},
	}
	root := f.root
	f.root = el
	for _, p := range cv.cells {
		f.put(p.X, p.Y, p.Cell)
	}
	f.root = root
	f.root.Children = append(f.root.Children, el)
}

func (f *XmlFile) createElementForContent(x, y int, val Content) {
	el := xml.Element{
		Name: xml.NewName("content"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("x"), strconv.Itoa(x)),
			xml.NewAttribute(xml.NewName("y"), strconv.Itoa(y)),
		},
		Children: []xml.Node{
			xml.NewText(string(val.Value)),
		},
	}
	f.root.Children = append(f.root.Children, &el)
}

func (f *XmlFile) createElementForConnector(x, y int, conn Connector) {
	el := &xml.Element{
		Name: xml.NewName("connector"),
	}
	for _, seg := range conn.Paths {
		sub := f.createElementForSegment(seg)
		el.Children = append(el.Children, sub)
	}
	f.root.Children = append(f.root.Children, el)
}

func (f *XmlFile) createElementForSegment(seg Segment) *xml.Element {
	start := &xml.Element{
		Name: xml.NewName("start"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("x"), strconv.Itoa(seg.Start.X)),
			xml.NewAttribute(xml.NewName("y"), strconv.Itoa(seg.Start.Y)),
		},
	}
	end := &xml.Element{
		Name: xml.NewName("end"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("x"), strconv.Itoa(seg.End.X)),
			xml.NewAttribute(xml.NewName("y"), strconv.Itoa(seg.End.Y)),
		},
	}

	return &xml.Element{
		Name: xml.NewName("segment"),
		Children: []xml.Node{
			start,
			end,
		},
	}
}

type Svg struct {
	root *svg.Document
	opts SvgOptions
}

func NewSvg(opts SvgOptions, size Dimension) (View, error) {
	doc := svg.NewDocument(float64(size.Width), float64(size.Height))
	return &Svg{
		root: doc,
		opts: opts,
	}, nil
}

func (s *Svg) Render(w io.Writer) error {
	return s.root.Render(w)
}

func (s *Svg) put(x, y int, cell Cell) error {
	var el svg.Element
	switch c := cell.(type) {
	case Content:
		el = svg.NewText(float64(x), float64(y), string(c.Value))
	case Connector:
		p := svg.NewPath()
		for i, s := range c.Paths {
			if i == 0 {
				p.MoveTo(float64(s.Start.X), float64(s.Start.Y))
			} else {
				p.LineTo(float64(s.Start.X), float64(s.Start.Y))
			}
			p.LineTo(float64(s.End.X), float64(s.End.Y))
		}
		el = p
	default:
	}
	s.root.Append(el)
	return nil
}

type Table struct{}

func NewTable() (View, error) {
	t := &Table{}
	return t, nil
}

func (t *Table) Render(w io.Writer) error {
	return nil
}

func (t *Table) put(x, y int, cell Cell) error {
	return nil
}

type Screen struct {
	lines [][]rune
	dim   Dimension

	opts ScreenOptions

	connector ConnectorStyle
	border    bool
	ticksStep int

	crossings []Point
}

func NewScreen(opts ScreenOptions, size Dimension) (View, error) {
	sc := &Screen{
		lines:     make([][]rune, size.Height),
		dim:       size,
		border:    opts.Border,
		connector: opts.Style,
		ticksStep: opts.CoordinatesStep,
	}
	if err := sc.dim.Validate(); err != nil {
		return nil, err
	}
	for i := range sc.lines {
		sc.lines[i] = make([]rune, size.Width)
	}
	sc.fillGrid()
	return sc, nil
}

func (s *Screen) Render(w io.Writer) error {
	var (
		ws     = bufio.NewWriter(w)
		spaces = strings.Repeat(" ", 6)
	)
	if s.showCoordinates() {
		if _, err := ws.WriteString(spaces); err != nil {
			return err
		}
		if s.border {
			if _, err := ws.WriteRune(space); err != nil {
				return err
			}
		}
		row := s.coordinatesX()
		for i := range row {
			if _, err := ws.WriteRune(row[i]); err != nil {
				return err
			}
		}
		if _, err := ws.WriteRune('\n'); err != nil {
			return err
		}
	}
	if s.border {
		if s.showCoordinates() {
			if _, err := ws.WriteString(spaces); err != nil {
				return err
			}
		}
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
		if s.showCoordinates() {
			if i%s.ticksStep == 0 {
				y := strconv.Itoa(i)
				if _, err := ws.WriteString(strings.Repeat(" ", 5-len(y))); err != nil {
					return err
				}
				if _, err := ws.WriteString(y + " "); err != nil {
					return err
				}
			} else {
				if _, err := ws.WriteString(spaces); err != nil {
					return err
				}
			}
		}
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
		if s.showCoordinates() {
			ws.WriteString(spaces)
		}
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

func (s *Screen) showCoordinates() bool {
	return s.ticksStep > 0
}

func (s *Screen) coordinatesX() []rune {
	line := make([]rune, s.dim.Width)
	for i := range line {
		line[i] = space
	}
	for i := 0; i < s.dim.Width; i += s.ticksStep {
		ix := strconv.Itoa(i)
		if i == 0 {
			copy(line[i:i+1], []rune(ix))
		} else {
			copy(line[i-len(ix):i], []rune(ix))
		}
	}
	return line
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
		// return IsBar(s.lines[y][x])
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
