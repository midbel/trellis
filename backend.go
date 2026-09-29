package trellis

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/midbel/angle/svg"
	"github.com/midbel/angle/xml"
	"github.com/midbel/curly"
)

const space = ' '

type drawer interface {
	draw(int, int, Cell) error
}

type View interface {
	Render(io.Writer) error
	fill(int, int, func(drawer) error) error
	drawer
}

type JsonFile struct {
	root map[string]any
	opts JsonOptions
}

func NewJson(opts JsonOptions) (View, error) {
	j := &JsonFile{
		root: make(map[string]any),
		opts: opts,
	}
	j.root["width"] = opts.Size.Width
	j.root["height"] = opts.Size.Height
	j.root["orientation"] = opts.Orient.String()
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

func (f *JsonFile) draw(x, y int, cell Cell) error {
	switch c := cell.(type) {
	case Content:
		v := createJsonContent(x, y, c)
		vs, ok := f.root["cells"].([]any)
		if ok {
			f.root["cells"] = append(vs, v)
		}
	case Connector:
		v := createJsonConnector(x, y, c)
		vs, ok := f.root["connectors"].([]any)
		if ok {
			f.root["connectors"] = append(vs, v)
		}
	default:
		return fmt.Errorf("element can not be draw")
	}
	return nil
}

func (f *JsonFile) fill(x, y int, fill func(drawer) error) error {
	j := newJsonElement(x, y)
	if err := fill(j); err != nil {
		return err
	}
	vs, ok := f.root["canvas"].([]any)
	if ok {
		f.root["canvas"] = append(vs, j.canvas)
	}
	return nil
}

type jsonElement struct {
	canvas map[string]any
}

func newJsonElement(x, y int) *jsonElement {
	e := &jsonElement{
		canvas: make(map[string]any),
	}
	e.canvas["x"] = x
	e.canvas["y"] = y
	e.canvas["cells"] = []any{}
	e.canvas["connectors"] = []any{}
	e.canvas["canvas"] = []any{}
	return e
}

func (e *jsonElement) draw(x, y int, cell Cell) error {
	switch c := cell.(type) {
	case Content:
		v := createJsonContent(x, y, c)
		vs, ok := e.canvas["cells"].([]any)
		if ok {
			e.canvas["cells"] = append(vs, v)
		}
	case Connector:
		v := createJsonConnector(x, y, c)
		vs, ok := e.canvas["connectors"].([]any)
		if ok {
			e.canvas["connectors"] = append(vs, v)
		}
	default:
		return fmt.Errorf("element can not be draw")
	}
	return nil
}

func createJsonContent(x, y int, c Content) any {
	v := map[string]any{
		"content": string(c.Value),
		"x":       x,
		"y":       y,
	}
	return v
}

func createJsonConnector(x, y int, c Connector) any {
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
	return list
}

type XmlFile struct {
	root *xml.Element
	opts XmlOptions
}

func NewXml(opts XmlOptions) (View, error) {
	el := xml.Element{
		Name: xml.NewName("canvas"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("width"), strconv.Itoa(opts.Size.Width)),
			xml.NewAttribute(xml.NewName("height"), strconv.Itoa(opts.Size.Height)),
			xml.NewAttribute(xml.NewName("orientation"), opts.Orient.String()),
		},
	}
	return &XmlFile{
		root: &el,
		opts: opts,
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

func (f *XmlFile) draw(x, y int, cell Cell) error {
	var node xml.Node
	switch c := cell.(type) {
	case Content:
		node = createElementForContent(x, y, c)
	case Connector:
		node = createElementForConnector(x, y, c)
	default:
	}
	if node != nil {
		f.root.Children = append(f.root.Children, node)
	}
	return nil
}

func (f *XmlFile) fill(x, y int, fill func(drawer) error) error {
	el := newXmlElement(x, y)
	if err := fill(el); err != nil {
		return err
	}
	f.root.Children = append(f.root.Children, el.root)
	return nil
}

type xmlElement struct {
	root *xml.Element
}

func newXmlElement(x, y int) *xmlElement {
	el := &xml.Element{
		Name: xml.NewName("canvas"),
		Attributes: []xml.Attribute{
			xml.NewAttribute(xml.NewName("x"), strconv.Itoa(x)),
			xml.NewAttribute(xml.NewName("y"), strconv.Itoa(y)),
		},
	}
	return &xmlElement{
		root: el,
	}
}

func (e *xmlElement) draw(x, y int, cell Cell) error {
	var node xml.Node
	switch c := cell.(type) {
	case Content:
		node = createElementForContent(x, y, c)
	case Connector:
		node = createElementForConnector(x, y, c)
	default:
	}
	if node != nil {
		e.root.Children = append(e.root.Children, node)
	}
	return nil
}

func createElementForContent(x, y int, val Content) xml.Node {
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
	return &el
}

func createElementForConnector(x, y int, conn Connector) xml.Node {
	el := &xml.Element{
		Name: xml.NewName("connector"),
	}
	for _, seg := range conn.Paths {
		sub := createElementForSegment(seg)
		el.Children = append(el.Children, sub)
	}
	return el
}

func createElementForSegment(seg Segment) xml.Node {
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

func NewSvg(opts SvgOptions) (View, error) {
	doc := svg.NewDocument(float64(opts.Size.Width), float64(opts.Size.Height))
	return &Svg{
		root: doc,
		opts: opts,
	}, nil
}

func (s *Svg) Render(w io.Writer) error {
	return s.root.Render(w)
}

func (s *Svg) draw(x, y int, cell Cell) error {
	var el svg.Element
	switch c := cell.(type) {
	case Content:
		el = svg.NewText(float64(x), float64(y), string(c.Value))
	case Connector:
		x, err := s.putConnector(c)
		if err != nil {
			return err
		}
		el = x
	default:
	}
	s.root.Append(el)
	return nil
}

func (s *Svg) fill(x, y int, fill func(drawer) error) error {
	return fill(s)
}

func (s *Svg) putConnector(c Connector) (svg.Element, error) {
	switch s.opts.Path {
	case ManathanPath:
		return s.manathanPath(c)
	case DirectPath:
		return s.directPath(c)
	case CurvePath:
		return s.curvePath(c)
	default:
		return nil, fmt.Errorf("unsupported path type")
	}
}

func (s *Svg) curvePath(c Connector) (svg.Element, error) {
	p := svg.NewPath()
	p.MoveTo(float64(c.X()), float64(c.Y()))
	if len(c.Paths) == 1 {
		p.LineTo(float64(c.Paths[0].End.X), float64(c.Paths[0].End.Y))
		return p, nil
	}
	if n := len(c.Paths) - 1; n > 0 {
		x := c.Paths[n].End.X
		y := c.Paths[n].End.Y
		d := x - c.X()
		p.CurveTo(
			float64(x),
			float64(y),
			float64(c.X()+d),
			float64(c.Y()),
			float64(x-d),
			float64(y),
		)
	}
	return p, nil
}

func (s *Svg) directPath(c Connector) (svg.Element, error) {
	p := svg.NewPath()
	p.MoveTo(float64(c.X()), float64(c.Y()))
	if len(c.Paths) == 1 {
		p.LineTo(float64(c.Paths[0].End.X), float64(c.Paths[0].End.Y))
		return p, nil
	}
	if n := len(c.Paths) - 1; n > 0 {
		p.LineTo(float64(c.Paths[n].End.X), float64(c.Paths[n].End.Y))
	}
	return p, nil
}

func (s *Svg) manathanPath(c Connector) (svg.Element, error) {
	p := svg.NewPath()
	for i, s := range c.Paths {
		if i == 0 {
			p.MoveTo(float64(s.Start.X), float64(s.Start.Y))
		} else {
			p.LineTo(float64(s.Start.X), float64(s.Start.Y))
		}
		p.LineTo(float64(s.End.X), float64(s.End.Y))
	}
	return p, nil
}

type Table struct {
	cells []Placement
}

func NewTable() (View, error) {
	t := &Table{}
	return t, nil
}

func (t *Table) Render(w io.Writer) error {
	wt := tabwriter.NewWriter(w, 0, 4, 2, '\t', 0)
	for _, p := range t.cells {
		var lines []string
		switch c := p.Cell.(type) {
		case Content:
			lines = t.getContentInfo(p.X, p.Y, c)
		case Connector:
			lines = t.getConnectorInfo(p.X, p.Y, c)
		default:
			return fmt.Errorf("element can not be rendered")
		}
		if len(lines) == 0 {
			continue
		}
		str := strings.Join(lines, "\t")
		if _, err := io.WriteString(wt, str + "\n"); err != nil {
			return err
		}
	}
	return wt.Flush()
}

func (t *Table) draw(x, y int, cell Cell) error {
	p := Placement{
		Point: NewPoint(x, y),
		Cell:  cell,
	}
	t.cells = append(t.cells, p)
	return nil
}

func (t *Table) fill(x, y int, fill func(drawer) error) error {
	return fill(t)
}

func (t *Table) getContentInfo(x, y int, c Content) []string {
	return []string{
		string(c.Value),
		strconv.Itoa(x),
		strconv.Itoa(y),
	}
}

func (t *Table) getConnectorInfo(x, y int, c Connector) []string {
	return []string{
		"connector",
		strconv.Itoa(c.X()),
		strconv.Itoa(c.Y()),
	}
}

type Screen struct {
	lines [][]rune
	dim   Dimension

	opts ScreenOptions

	crossings []Point
}

func NewScreen(opts ScreenOptions) (View, error) {
	if err := opts.Size.Validate(); err != nil {
		return nil, err
	}
	sc := &Screen{
		lines: make([][]rune, opts.Size.Height),
		dim:   opts.Size,
		opts:  opts,
	}
	for i := range sc.lines {
		sc.lines[i] = make([]rune, opts.Size.Width)
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
		if s.showBorder() {
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
	if s.showBorder() {
		if s.showCoordinates() {
			if _, err := ws.WriteString(spaces); err != nil {
				return err
			}
		}
		if _, err := ws.WriteRune(s.connector().TopLeft()); err != nil {
			return err
		}
		for range s.dim.Width {
			if _, err := ws.WriteRune(s.connector().HorizontalBar()); err != nil {
				return err
			}
		}
		if _, err := ws.WriteRune(s.connector().TopRight()); err != nil {
			return err
		}
		if _, err := ws.WriteRune('\n'); err != nil {
			return err
		}
	}
	s.writeCrossings()
	for i := range s.lines {
		if s.showCoordinates() {
			if i%s.opts.CoordinatesStep == 0 {
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
		if s.showBorder() {
			if _, err := ws.WriteRune(s.connector().VerticalBar()); err != nil {
				return err
			}
		}
		for j := range s.lines[i] {
			_, err := ws.WriteRune(s.lines[i][j])
			if err != nil {
				return err
			}
		}
		if s.showBorder() {
			if _, err := ws.WriteRune(s.connector().VerticalBar()); err != nil {
				return err
			}
		}
		if _, err := ws.WriteRune('\n'); err != nil {
			return err
		}
	}
	if s.showBorder() {
		if s.showCoordinates() {
			ws.WriteString(spaces)
		}
		if _, err := ws.WriteRune(s.connector().BottomLeft()); err != nil {
			return err
		}
		for range s.dim.Width {
			if _, err := ws.WriteRune(s.connector().HorizontalBar()); err != nil {
				return err
			}
		}
		if _, err := ws.WriteRune(s.connector().BottomRight()); err != nil {
			return err
		}
		if _, err := ws.WriteRune('\n'); err != nil {
			return err
		}
	}
	return ws.Flush()
}

func (s *Screen) draw(x, y int, cell Cell) error {
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

func (s *Screen) fill(x, y int, fill func(drawer) error) error {
	return fill(s)
}

func (s *Screen) writeCrossings() {
	isBar := func(y, x int) bool {
		if !s.dim.Valid(x, y) {
			return false
		}
		return IsBar(s.lines[y][x])
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
			char = s.connector().CrossPath()
		case left && right && top && !bottom:
			// horizontal up
			char = s.connector().HorizontalUp()
		case left && right && bottom && !top:
			// horizontal down
			char = s.connector().HorizontalDown()
		case top && bottom && left && !right:
			// vertical left
			char = s.connector().VerticalLeft()
		case top && bottom && right && !left:
			// vertical right
			char = s.connector().VerticalRight()
		case bottom && left && !right && !top:
			// bottom left
			char = s.connector().TopRight()
		case top && left && !right && !bottom:
			// top left
			char = s.connector().BottomRight()
		case bottom && right && !left && !top:
			// bottom right
			char = s.connector().TopLeft()
		case top && right && !left && !bottom:
			// top right
			char = s.connector().BottomLeft()
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
	char := s.connector().VerticalBar()
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
	char := s.connector().HorizontalBar()
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

func (s *Screen) connector() ConnectorStyle {
	return s.opts.Style
}

func (s *Screen) showBorder() bool {
	return s.opts.Border
}

func (s *Screen) showCoordinates() bool {
	return s.opts.CoordinatesStep > 0
}

func (s *Screen) coordinatesX() []rune {
	line := make([]rune, s.dim.Width)
	for i := range line {
		line[i] = space
	}
	for i := 0; i < s.dim.Width; i += s.opts.CoordinatesStep {
		ix := strconv.Itoa(i)
		if i == 0 {
			copy(line[i:i+1], []rune(ix))
		} else {
			copy(line[i-len(ix):i], []rune(ix))
		}
	}
	return line
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
