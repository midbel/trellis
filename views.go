package trellis

import (
	"fmt"
	"io"
	"strings"
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

type TableOptions struct {}

func (*TableOptions) Format() Output {
	return OutputTable
}

func (*TableOptions) applyDefaults() {}

type Table struct {
	cells []Placement
}

func NewTable() (View, error) {
	t := &Table{}
	return t, nil
}

func (t *Table) Render(w io.Writer) error {
	if err := t.renderHeader(w); err != nil {
		return err
	}
	for _, p := range t.cells {
		var lines []any
		switch c := p.Cell.(type) {
		case Content:
			lines = t.getContentInfo(p.X, p.Y, c)
		case Path:
			lines = t.getPathInfo(p.X, p.Y, c)
		default:
			return fmt.Errorf("invalid cell type")
		}
		if len(lines) == 0 {
			continue
		}
		if err := t.renderLine(w, lines); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) renderLine(w io.Writer, lines []any) error {
	_, err := fmt.Fprintf(w, "%-16s | %6d | %6d", lines...)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w)
	return err
}

func (t *Table) renderHeader(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%-16s | %6s | %6s", "Value", "X", "Y")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w)
	return err
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

func (t *Table) getContentInfo(x, y int, c Content) []any {
	return []any{
		strings.TrimSpace(string(c.Value)),
		x,
		y,
	}
}

func (t *Table) getPathInfo(x, y int, c Path) []any {
	return []any{
		"path",
		c.X(),
		c.Y(),
	}
}
