package trellis

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ScreenOptions struct {
	RenderOptions
	Border          bool
	CoordinatesStep int
	Style           ConnectorStyle
}

func (o *ScreenOptions) Layout() (RenderOptions, error) {
	var err error
	if o.Path != ManatthanPath {
		err = fmt.Errorf("manatthan default for screen only")
	}
	if err == nil {
		err = o.RenderOptions.Validate()
	}
	return o.RenderOptions, err
}

func (*ScreenOptions) Format() Output {
	return OutputScreen
}

func (o *ScreenOptions) applyDefaults() {
	o.RenderOptions.applyDefaults()
	if o.Style == 0 {
		o.Style = ConnectorUnicode
	}
}

type styleExtent struct {
	Style
	Length int
}

type Screen struct {
	lines  [][]rune
	styles map[Point]styleExtent
	dim    Dimension

	opts ScreenOptions

	crossings []Point
}

func NewScreen(opts ScreenOptions) (View, error) {
	if err := opts.Size.Validate(); err != nil {
		return nil, err
	}
	sc := &Screen{
		lines:  make([][]rune, opts.Size.Height),
		styles: make(map[Point]styleExtent),
		dim:    opts.Size,
		opts:   opts,
	}
	for i := range sc.lines {
		sc.lines[i] = make([]rune, opts.Size.Width)
	}
	sc.initGrid()
	return sc, nil
}

func (s *Screen) Render(w io.Writer) error {
	var (
		ws     = bufio.NewWriter(w)
		spaces = strings.Repeat(" ", 6)
	)
	if err := s.writeCoordinates(ws, spaces); err != nil {
		return err
	}
	if err := s.writeHeader(ws, spaces); err != nil {
		return err
	}
	s.writeCrossings()
	for i := range s.lines {
		if err := s.writeLinePrefix(ws, i, spaces); err != nil {
			return err
		}
		if err := s.writeLine(ws, i); err != nil {
			return err
		}
	}
	if err := s.writeFooter(ws, spaces); err != nil {
		return err
	}
	return ws.Flush()
}

func (s *Screen) writeHeader(ws *bufio.Writer, spaces string) error {
	if !s.showBorder() {
		return nil
	}
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
	return nil
}

func (s *Screen) writeFooter(ws *bufio.Writer, spaces string) error {
	if !s.showBorder() {
		return nil
	}
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
	return nil
}

func (s *Screen) writeCoordinates(ws *bufio.Writer, spaces string) error {
	if !s.showCoordinates() {
		return nil
	}
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
	return nil
}

func (s *Screen) writeLinePrefix(ws *bufio.Writer, i int, spaces string) error {
	if !s.showCoordinates() {
		return nil
	}
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
	return nil
}

func (s *Screen) writeLine(ws *bufio.Writer, i int) error {
	if s.showBorder() {
		if _, err := ws.WriteRune(s.connector().VerticalBar()); err != nil {
			return err
		}
	}

	var (
		count  int
		length int
		open   bool
	)
	for j := range s.lines[i] {
		if st, ok := s.styles[NewPoint(j, i)]; ok {
			count = 0
			length = st.Length
			if err := writeAnsiStyle(ws, st.Color, st.IsBold(), st.IsItalic()); err != nil {
				return err
			}
			open = true
		}
		_, err := ws.WriteRune(s.lines[i][j])
		if err != nil {
			return err
		}
		count += RuneWidth(s.lines[i][j])
		if count == length && open {
			open = false
			count = 0
			if err := writeAnsiClose(ws); err != nil {
				return err
			}
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
	return nil
}

func (s *Screen) initGrid() {
	for i := range s.lines {
		for j := range s.lines[i] {
			s.lines[i][j] = space
		}
	}
}

func (s *Screen) draw(x, y int, cell Cell) error {
	var err error
	if !s.dim.Valid(x, y) {
		return fmt.Errorf("invalid coordinates (%d, %d)", x, y)
	}
	switch c := cell.(type) {
	case Content:
		err = s.createContent(x, y, c)
	case Path:
		err = s.createPath(x, y, c)
	default:
		err = fmt.Errorf("invalid cell type")
	}
	return err
}

func (s *Screen) fill(x, y int, fill func(drawer) error) error {
	return fill(s)
}

func (s *Screen) createContent(x, y int, val Content) error {
	var size int
	for _, r := range val.Value {
		s.writeChar(x, y, r)
		z := RuneWidth(r)
		x += z
		size += z
	}
	if !val.Style.Zero() {
		s.styles[NewPoint(x-size, y)] = styleExtent{
			Style:  val.Style,
			Length: size,
		}
	}
	return nil
}

func (s *Screen) createPath(x, y int, path Path) error {
	if s.opts.Path != ManatthanPath {
		return fmt.Errorf("screen only supports manatthan path")
	}
	var paths []Path
	if s.opts.Orient == HorizontalLayout {
		paths = splitPathH(path)
	} else {
		paths = splitPathV(path)
	}
	for i := range paths {
		if paths[i].Horizontal() {
			s.horizontalConnector(paths[i])
		} else {
			s.verticalConnector(paths[i])
		}
	}
	return nil
}

func (s *Screen) verticalConnector(path Path) {
	s.crossings = append(s.crossings, path.Start, path.End)

	char := s.connector().VerticalBar()
	if path.DistanceY() == 1 {
		s.writeSymbol(path.Start.X, path.Start.Y, char)
		return
	}
	var (
		start = path.Start
		end   = path.End
	)
	if start.BeforeY(end) {
		start, end = end, start
	}
	for y := start.Y; y >= end.Y; y-- {
		s.writeSymbol(start.X, y, char)
	}
}

func (s *Screen) horizontalConnector(path Path) {
	s.crossings = append(s.crossings, path.Start, path.End)

	char := s.connector().HorizontalBar()
	if path.DistanceX() == 1 {
		s.writeSymbol(path.Start.X, path.Start.Y, char)
		return
	}
	var (
		start = path.Start
		end   = path.End
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
